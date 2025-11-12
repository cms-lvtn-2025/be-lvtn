import { Job, Queue, Worker } from "bullmq";
import { IService, ServiceModel } from "../../../database";
import { ExternalService } from "../baseExternal";
import axios from "axios";
import { redisConnection, ServiceJobData } from "../../queue";


export class HttpService extends ExternalService {
  public name = "http";
  public version = "v.0.0.1";

  constructor(service: IService) {
    super(service);
  }

  

  public async createClient(): Promise<any> {
    this.client = axios.create({
      baseURL: this.service.url,
      headers: {
        'Content-Type': 'application/json',
      },
    });
    return this.client;
  }
  public async updateClient(): Promise<any> {
    return this.createClient();
  }
  public async deleteClient(): Promise<any> {
    this.client.close();
  }
  public async ping(): Promise<boolean> {
    try {
      const response = await this.client.get('/ping');
      return response.status === 200;
    } catch (error: any) {
      console.error(`${this.service.name} ping failed:`, error.message);
      return false;
    }
  }

  public async healthCheckAndUpdateClient(): Promise<boolean> {
    const isHealthy = await this.ping();
    if (isHealthy) {
      await ServiceModel.updateOne(
        { _id: this.service._id },
        {
          $set: {
            healthy: isHealthy,
            lastHealthCheck: new Date(),
          },
        }
      );
      return true;
    } else {
      console.error(`${this.service.name} healthCheckAndUpdateClient failed`);
      await ServiceModel.updateOne(
        { _id: this.service._id },
        {
          $set: {
            healthy: false,
            lastHealthCheck: new Date(),
          },
        }
      );
      return false;
    }
  }

  private async createQueueWorker(): Promise<any> {
    const worker = new Worker<ServiceJobData>(this.service.name, async (job: Job<ServiceJobData>) => {
      console.log(`${this.service.name} processing job ${job.id}`);
      console.log(`${this.service.name} job data:`, job.data);
      try {
        const response = await this.client.post(job.data.method, job.data.params);
        return response.data;
      } catch (error: any) {
        console.error(`${this.service.name} processing job ${job.id} failed:`, error.message);
        throw error;
      }
    }, {
      connection: redisConnection,
      concurrency: parseInt(process.env.WORKER_CONCURRENCY || "3"),
    });
    return worker;
  }

  private async eventWorker(): Promise<any> {
    this.worker?.on("completed", (job) => {
      console.log(`${this.service.name} job ${job?.id} completed`);
    });
    this.worker?.on("failed", (job, error) => {
      console.error(`${this.service.name} job ${job?.id} failed:`, error.message);
    });
  }

  public async createQueue(): Promise<any> {
    const check = await this.healthCheckAndUpdateClient();
    if (!check) {
      throw new Error(`${this.service.name} client not found`);
    }
    this.queue = this.queue ? await this.deleteQueue() : undefined;
    this.queue = new Queue<ServiceJobData>(this.service.name, {
      connection: redisConnection,
      defaultJobOptions: {
        attempts: 3,
        backoff: {
          type: "exponential",
          delay: 5000,
        },
        removeOnComplete: {
          count: 100,
        },
        removeOnFail: {
          count: 50,
        },
      },
    });
    await this.createQueueWorker();
    await this.eventWorker();
    return this.queue;
  }
  public async healthCheckAndUpdateQueue(): Promise<boolean> {
    const check = await this.healthCheckAndUpdateClient();
    if (!check) {
      throw new Error(`${this.service.name} client not found`);
    }
    return true;
  }
  public async pingQueue(): Promise<boolean> {
    const client = this.queue?.client;
    return client?.then((client) => client.status === "ready") || false;
  }
  public async deleteQueue(): Promise<any> {
    this.queue?.close();
    this.queue = undefined;
    return this.queue;
  }
  public async getQueue(): Promise<any> {
    return await this.queue;
  }
  public async updateQueue(): Promise<any> {
    this.queue?.close();
    this.queue = await this.createQueue();
    return this.queue;
  }
}