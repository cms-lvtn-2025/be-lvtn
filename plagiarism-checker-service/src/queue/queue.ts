import {
  Job,
  FlowProducer,
  FlowJob,
  Queue,
  Worker,
  JobsOptions,
  FlowOpts,
} from "bullmq";
import IORedis from "ioredis";
import { IService, IWorkflow, IWorkflowModel } from "../database/models";
import { loadGrpcClient } from "./grpc/client-loader";
import vm from "node:vm";
// Redis connection
import { v4 as uuidv4 } from "uuid";

const redisConnection = new IORedis({
  host: process.env.REDIS_HOST || "localhost",
  port: parseInt(process.env.REDIS_PORT || "10002"),
  password: process.env.REDIS_PASSWORD || undefined,
  db: parseInt(process.env.REDIS_DB || "0"),
  maxRetriesPerRequest: null,
});

export interface ServiceJobData {
  method: string;
  params: any;
  metadata?: any;
}

export interface ServiceQueueInfo {
  serviceName: string;
  queueName: string;
  queue: Queue<ServiceJobData>;
  worker: Worker<ServiceJobData>;
  client: any; // gRPC client hoặc class instance
  type: "dynamic" | "static"; // dynamic = từ DB, static = manual register
  service?: IService; // Optional - chỉ có khi type = dynamic
}

export interface Children {
  serviceName: string;
  method: string;
  params: any;
  options?: JobsOptions;
  children?: Children[];
}

export interface FlowJobWithId {
  id: string;
  flow: FlowJob;
}

/**
 * Service Queue Manager - Quản lý queue động cho mỗi service
 */
export class ServiceQueueManager {
  private queues: Map<string, ServiceQueueInfo> = new Map();

  /**
   * Tạo queue và worker cho một service
   */
  createServiceQueue(service: IService): ServiceQueueInfo {
    const queueName = `${service.name}`;

    console.log(`\n🔧 Creating queue for ${service.name}...`);
    // Load gRPC client
    const client = loadGrpcClient(service);

    // Tạo Queue
    const queue = new Queue<ServiceJobData>(queueName, {
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

    // Tạo Worker với gRPC client
    const worker = new Worker<ServiceJobData>(
      queueName,
      async (job: Job<ServiceJobData>) => {
        console.log(`\n⚡ [${service.name}] Processing job ${job.id}`);
        console.log(`   Method: ${job.data.method}`);
        console.log(`   Params:`, job.data.params);

        try {
          // Gọi gRPC method động
          const method = client[job.data.method];
          if (!method) {
            throw new Error(
              `Method ${job.data.method} not found on ${service.name}`
            );
          }
          const children = await job.getChildrenValues();

          if (typeof job.data.params == "object") {
            Object.entries(job.data.params).forEach(([key, value]) => {
              if (typeof value == "string" && value.startsWith("@bull:")) {
                // Parse: @bull:file_service-queue:jobId.file.id
                const cleaned = value.replace("@", ""); // bull:file_service-queue:jobId.file.id
                const [fullJobKey, ...pathParts] = cleaned.split("."); // ["bull:file_service-queue:jobId", "file", "id"]

                // Lấy child result từ children object
                let dataKey = children[fullJobKey]; // children["bull:file_service-queue:jobId"]

                // Navigate qua path (file.id)
                pathParts.forEach((part) => {
                  if (dataKey && typeof dataKey === "object") {
                    dataKey = dataKey[part];
                  }
                });

                job.data.params[key] = dataKey;
                console.log(`   🔄 Resolved ${value} -> ${dataKey}`);
              }
            });
          }
          console.log("   📦 Children results:", children);

          // Call gRPC method
          const result = await new Promise((resolve, reject) => {
            method.call(
              client,
              job.data.params,
              (error: any, response: any) => {
                if (error) {
                  reject(error);
                } else {
                  resolve(response);
                }
              }
            );
          });

          console.log(`   ✅ Success:`, result);
          return result;
        } catch (error: any) {
          console.error(`   ❌ Error:`, error.message);
          throw error;
        }
      },
      {
        connection: redisConnection,
        concurrency: parseInt(process.env.WORKER_CONCURRENCY || "3"),
      }
    );

    // Event listeners
    worker.on("completed", (job) => {
      console.log(`✨ [${service.name}] Job ${job.id} completed`);
    });

    worker.on("failed", (job, err) => {
      console.error(`❌ [${service.name}] Job ${job?.id} failed:`, err.message);
    });

    const queueInfo: ServiceQueueInfo = {
      serviceName: service.name,
      queueName,
      queue,
      worker,
      client,
      type: "dynamic",
      service,
    };

    this.queues.set(service.name, queueInfo);

    console.log(`✅ Queue created for ${service.name}`);

    return queueInfo;
  }
  /**
   * Đăng ký với service mongodb
   */
  createServiceWorkflowQueue(WorkflowModel: IWorkflowModel): ServiceQueueInfo {
    const queueName = `MONGODB_WORKFLOW`;

    const queue = new Queue<ServiceJobData>(queueName, {
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

    
    // Tạo Worker với MongoDB model
    const worker = new Worker<ServiceJobData>(
      queueName,
      async (job: Job<ServiceJobData>) => {
        console.log(`\n⚡ [MONGODB] Processing job ${job.id}`);
        console.log(`   Method: ${job.data.method}`);
        console.log(`   Params:`, job.data.params);

        try {
          // Gọi method từ WorkflowModel
          const method = (WorkflowModel as any)[job.data.method];
          if (!method) {
            throw new Error(
              `Method ${job.data.method} not found on MONGODB_WORKFLOW_QUEUE`
            );
          }

          // Call method (bind this context)
          const result = await method.call(WorkflowModel, job.data.params);

          console.log(`   ✅ Success:`, result);
          return result;
        } catch (error: any) {
          console.error(`   ❌ Error:`, error.message);
          throw error;
        }
      },
      {
        connection: redisConnection,
        concurrency: parseInt(process.env.WORKER_CONCURRENCY || "3"),
      }
    );
    const queueInfo: ServiceQueueInfo = {
      serviceName: "MONGODB_WORKFLOW",
      queueName,
      queue,
      worker,
      client: WorkflowModel,
      type: "static",
    };
    this.queues.set("MONGODB_WORKFLOW", queueInfo);
    console.log(`✅ Queue created for MONGODB_WORKFLOW`);
    return queueInfo;
  }

  /**
   * Đăng ký static queue với class instance
   * @param serviceName - Tên service (uppercase)
   * @param serviceInstance - Instance của class service
   * @param options - Queue options
   */
  registerStaticQueue(
    serviceName: string,
    serviceInstance: any,
    options?: {
      concurrency?: number;
      attempts?: number;
    }
  ): ServiceQueueInfo {
    const queueName = `${serviceName}`;

    console.log(`\n🔧 Registering static queue for ${serviceName}...`);

    // Tạo Queue
    const queue = new Queue<ServiceJobData>(queueName, {
      connection: redisConnection,
      defaultJobOptions: {
        attempts: options?.attempts || 3,
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

    // Tạo Worker với service instance
    const worker = new Worker<ServiceJobData>(
      queueName,
      async (job: Job<ServiceJobData>) => {
        console.log(`\n⚡ [${serviceName}] Processing job ${job.id}`);
        console.log(`   Method: ${job.data.method}`);
        console.log(`   Params:`, job.data.params);
        const children = await job.getChildrenValues();

        try {
          console.log("   📦 Children results:", children);
          if (job.data.method == "EnJob") {
            return children;
          }
          // Gọi method từ service instance
          const method = serviceInstance[job.data.method];
          if (!method || typeof method !== "function") {
            throw new Error(
              `Method ${job.data.method} not found on ${serviceName}`
            );
          }
          if (typeof job.data.params == "object") {
            Object.entries(job.data.params).forEach(([key, value]) => {
              if (typeof value == "string" && value.startsWith("@bull:")) {
                // Parse: @bull:file_service-queue:jobId.file.id
                const cleaned = value.replace("@", ""); // bull:file_service-queue:jobId.file.id
                const [fullJobKey, ...pathParts] = cleaned.split("."); // ["bull:file_service-queue:jobId", "file", "id"]

                // Lấy child result từ children object
                let dataKey = children[fullJobKey]; // children["bull:file_service-queue:jobId"]

                // Navigate qua path (file.id)
                pathParts.forEach((part) => {
                  if (dataKey && typeof dataKey === "object") {
                    dataKey = dataKey[part];
                  }
                });

                job.data.params[key] = dataKey;
                console.log(`   🔄 Resolved ${value} -> ${dataKey}`);
              }
            });
          }

          // Call method (bind this context)
          const result = await method.call(serviceInstance, job.data.params);

          console.log(`   ✅ Success:`, result);
          return result;
        } catch (error: any) {
          console.error(`   ❌ Error:`, error.message);
          throw error;
        }
      },
      {
        connection: redisConnection,
        concurrency:
          options?.concurrency ||
          parseInt(process.env.WORKER_CONCURRENCY || "3"),
      }
    );

    // Event listeners
    worker.on("completed", (job) => {
      console.log(`✨ [${serviceName}] Job ${job.id} completed`);
    });

    worker.on("failed", (job, err) => {
      console.error(`❌ [${serviceName}] Job ${job?.id} failed:`, err.message);
    });

    const queueInfo: ServiceQueueInfo = {
      serviceName: serviceName.toUpperCase(),
      queueName,
      queue,
      worker,
      client: serviceInstance,
      type: "static",
    };

    this.queues.set(serviceName.toUpperCase(), queueInfo);

    console.log(`✅ Static queue created for ${serviceName}`);

    return queueInfo;
  }

  /**
   * Lấy queue info của một service
   */
  getServiceQueue(serviceName: string): ServiceQueueInfo | undefined {
    return this.queues.get(serviceName);
  }

  /**
   * Lấy tất cả queues
   */
  getAllQueues(): ServiceQueueInfo[] {
    return Array.from(this.queues.values());
  }

  /**
   * Thêm job vào queue của service
   */
  async addJob(
    serviceName: string,
    data: ServiceJobData
  ): Promise<Job<ServiceJobData>> {
    const queueInfo = this.queues.get(serviceName);
    if (!queueInfo) {
      throw new Error(`Queue for service ${serviceName} not found`);
    }

    const job = await queueInfo.queue.add("grpc-call", data);
    console.log(`📝 Added job ${job.id} to ${serviceName} queue`);
    return job;
  }

  /**
   * Đóng tất cả queues và workers
   */
  async closeAll(): Promise<void> {
    console.log("\n🛑 Closing all queues...");

    for (const [name, info] of this.queues.entries()) {
      await info.queue.close();
      await info.worker.close();
      console.log(`  ✅ Closed ${name} queue`);
    }

    await redisConnection.quit();
    console.log("✅ All queues closed");
  }
}

// Singleton instance
export const serviceQueueManager = new ServiceQueueManager();

/**
 * ==========================================
 * Job Management Service - Quản lý job workflows
 * ==========================================
 */
export class QueueService {
  private flowProducer: FlowProducer;

  constructor() {
    // Flow Producer để tạo job hierarchies
    this.flowProducer = new FlowProducer({
      connection: {
        host: process.env.REDIS_HOST || "localhost",
        port: parseInt(process.env.REDIS_PORT || "10002"),
        password: process.env.REDIS_PASSWORD || undefined,
        db: parseInt(process.env.REDIS_DB || "0"),
      },
    });
  }

  /**
   * Tạo một job đơn giản
   */
  async createJob(
    serviceName: string,
    method: string,
    params: any,
    _options?: {
      jobId?: string;
      delay?: number;
      priority?: number;
      attempts?: number;
    }
  ): Promise<Job<ServiceJobData>> {
    const jobData: ServiceJobData = {
      method,
      params,
      metadata: {
        serviceName,
        createdAt: new Date().toISOString(),
      },
    };

    const job = await serviceQueueManager.addJob(serviceName, jobData);

    console.log(`📝 Created job ${job.id} for ${serviceName}.${method}`);
    return job;
  }

  async evaluateJob(params: any): Promise<any> {
    console.log(
      "params in evaluateJob:",
      params,
      typeof params,
      params.code,
      params.returnValue
    );
    if (typeof params === "object" && params.code && params.returnValue) {
      const { v4: uuidv4 } = require("uuid");
      const sandbox = {
        returnValue: params.returnValue,
        data: params.data || {},
        console,
        createJobWithChildren: this.createJobWithChildren.bind(this),
        uuidv4,
        Date,
      };

      const script = new vm.Script(`
      (async () => {
        ${params.code}
      })()
    `);

      try {
        const result = await script.runInNewContext(sandbox);
        return result;
      } catch (err) {
        console.error("Error evaluating code:", err);
        throw err;
      }
    } else {
      throw new Error(
        "Invalid params: must include code, data, and returnValue"
      );
    }
  }

  /**
   * Params add id of child job
   */
  public async addIdOfChildJob(params: any, flowJobWithId: FlowJobWithId[]) {
    if (typeof params === "object") {
      Object.entries(params).forEach(([key, value]) => {
        // value have "@__id__{index}:..."
        if (typeof value == "string" && value.startsWith("@__id__")) {
          const match = value.match(/^@__id__(\d+):(.*)$/);
          console.log(match)
          if (match) {
            const index = parseInt(match[1], 10);
            let dataKey = flowJobWithId[index];
            params[key] = `@bull:${dataKey.flow.queueName}:${dataKey.id}${match[2]?'.':''}${match[2]}`;
          }
        }else if (typeof value == "object") {
          this.addIdOfChildJob(value, flowJobWithId);
        }
      });
    }
  }

  /**
   * Tạo job với child jobs (parent-child relationship)
   * Parent job chờ tất cả child jobs hoàn thành
   */
  /**
   * Helper function để tạo FlowJob đệ quy (support nested children)
   */
  private buildFlowJob(
    child: Children,
    index: number,
    parentServiceName?: string
  ): FlowJobWithId {
    const queueName = `${child.serviceName}`;

    // Đệ quy xử lý children của child (grandchildren)
    const grandChildren: FlowJobWithId[] | undefined = child.children
      ? child.children.map((grandChild, idx) =>
          this.buildFlowJob(grandChild, idx, child.serviceName)
        )
      : undefined;
    const id = uuidv4();
    console.log(child.params);
    this.addIdOfChildJob(child.params, grandChildren || []);
    console.log("After addIdOfChildJob:", child.params);
    return {
      id: id,
      flow: {
        name: `child-${index}`,
        queueName,
        data: {
          method: child.method,
          params: child.params,
          metadata: {
            serviceName: child.serviceName,
            parentService: parentServiceName,
            hasChildren: !!grandChildren,
          },
        },
        opts: {
          jobId: id,
          ...child.options,
        },
        children: grandChildren?.map((value) => value.flow), // ✅ Nested children support
      },
    };
  }

  async createJobWithChildren(
    parentServiceName: string,
    parentMethod: string,
    parentParams: any,
    children: Children[],
    options?: JobsOptions,
    FlowOpts?: FlowOpts
  ): Promise<any> {
    const queueName = `${parentServiceName}`;

    // Tạo child jobs với đệ quy support
    const childJobs: FlowJobWithId[] = children.map((child, index) =>
      this.buildFlowJob(child, index, parentServiceName)
    );

    // Tạo parent job với children
    const flow = await this.flowProducer.add(
      {
        name: "parent",
        queueName,
        data: {
          method: parentMethod,
          params: parentParams,
          metadata: {
            serviceName: parentServiceName,
            hasChildren: true,
          },
        },
        children: childJobs.map((value) => value.flow),
        opts: {
          ...options,
        },
      },
      FlowOpts
    );

    // Đếm tổng số jobs (bao gồm cả nested)
    const countJobs = (children: Children[]): number => {
      return children.reduce((total, child) => {
        return total + 1 + (child.children ? countJobs(child.children) : 0);
      }, 0);
    };
    const totalJobs = countJobs(children);

    console.log(
      `📝 Created parent job with ${children.length} direct children (${totalJobs} total jobs) for ${parentServiceName}.${parentMethod}`
    );
    return flow;
  }

  /**
   * Get job với children results
   */
  async getJobWithChildren(serviceName: string, jobId: string): Promise<any> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    const job = await queueInfo.queue.getJob(jobId);
    if (!job) {
      return null;
    }

    const state = await job.getState();
    const childrenValues = await job.getChildrenValues();

    // Parse children results
    const children = Object.entries(childrenValues).map(([jobId, result]) => ({
      jobId,
      result,
    }));

    return {
      id: job.id,
      name: job.name,
      data: job.data,
      state,
      returnValue: job.returnvalue,
      children,
      childrenCount: children.length,
    };
  }

  /**
   * Get children results only (helper)
   */
  async getChildrenResults(serviceName: string, jobId: string): Promise<any[]> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    const job = await queueInfo.queue.getJob(jobId);
    if (!job) {
      return [];
    }

    const childrenValues = await job.getChildrenValues();

    // Trả về array của results
    return Object.values(childrenValues);
  }

  /**
   * Get job status
   */
  async getJobStatus(serviceName: string, jobId: string): Promise<any> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    const job = await queueInfo.queue.getJob(jobId);
    if (!job) {
      return null;
    }

    const state = await job.getState();
    const progress = job.progress;
    const returnValue = job.returnvalue;
    const failedReason = job.failedReason;

    return {
      id: job.id,
      name: job.name,
      data: job.data,
      state,
      progress,
      returnValue,
      failedReason,
      attemptsMade: job.attemptsMade,
      timestamp: job.timestamp,
      processedOn: job.processedOn,
      finishedOn: job.finishedOn,
    };
  }

  /**
   * Cancel/Remove a job
   */
  async cancelJob(serviceName: string, jobId: string): Promise<void> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    const job = await queueInfo.queue.getJob(jobId);
    if (job) {
      await job.remove();
      console.log(`🗑️  Cancelled job ${jobId} from ${serviceName}`);
    }
  }

  /**
   * Retry a failed job
   */
  async retryJob(serviceName: string, jobId: string): Promise<void> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    const job = await queueInfo.queue.getJob(jobId);
    if (job) {
      await job.retry();
      console.log(`🔄 Retrying job ${jobId} from ${serviceName}`);
    }
  }

  /**
   * Get queue statistics
   */
  async getQueueStats(serviceName: string): Promise<any> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    const queue = queueInfo.queue;

    const [waiting, active, completed, failed, delayed] = await Promise.all([
      queue.getWaitingCount(),
      queue.getActiveCount(),
      queue.getCompletedCount(),
      queue.getFailedCount(),
      queue.getDelayedCount(),
    ]);

    return {
      serviceName,
      waiting,
      active,
      completed,
      failed,
      delayed,
      total: waiting + active + completed + failed + delayed,
    };
  }

  /**
   * Get all queues stats
   */
  async getAllQueuesStats(): Promise<any[]> {
    const allQueues = serviceQueueManager.getAllQueues();

    const stats = await Promise.all(
      allQueues.map(async (queueInfo) => {
        return this.getQueueStats(queueInfo.serviceName);
      })
    );

    return stats;
  }

  /**
   * Pause a queue
   */
  async pauseQueue(serviceName: string): Promise<void> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    await queueInfo.queue.pause();
    console.log(`⏸️  Paused queue for ${serviceName}`);
  }

  /**
   * Resume a queue
   */
  async resumeQueue(serviceName: string): Promise<void> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    await queueInfo.queue.resume();
    console.log(`▶️  Resumed queue for ${serviceName}`);
  }

  /**
   * Clean completed jobs
   */
  async cleanCompletedJobs(
    serviceName: string,
    olderThanMs: number = 3600000
  ): Promise<void> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    await queueInfo.queue.clean(olderThanMs, 100, "completed");
    console.log(`🧹 Cleaned completed jobs for ${serviceName}`);
  }

  /**
   * Clean failed jobs
   */
  async cleanFailedJobs(
    serviceName: string,
    olderThanMs: number = 3600000
  ): Promise<void> {
    const queueInfo = serviceQueueManager.getServiceQueue(serviceName);
    if (!queueInfo) {
      throw new Error(`Service ${serviceName} not found`);
    }

    await queueInfo.queue.clean(olderThanMs, 100, "failed");
    console.log(`🧹 Cleaned failed jobs for ${serviceName}`);
  }

  /**
   * Close flow producer
   */
  async close(): Promise<void> {
    await this.flowProducer.close();
  }
}

// Singleton instance
export const queueService = new QueueService();
