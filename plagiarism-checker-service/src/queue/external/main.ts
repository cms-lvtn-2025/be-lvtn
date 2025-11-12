import { IService } from "../../database";
import { GrpcService } from "./grpc/main";
import { ExternalService } from "./baseExternal";
import { Worker } from "bullmq";
import { ServiceJobData } from "../queue";




export class ManagerExternalService {
  private externalServices: Map<IService, ExternalService>;

  constructor() {
    this.externalServices = new Map();
  }

  public async getWorker(service: IService): Promise<Worker<ServiceJobData, any, string>>  {
    const externalService = this.externalServices.get(service);
    if (!externalService) {
      throw new Error(`Service ${service.name} not found`);
    }
    return externalService.worker as Worker<ServiceJobData, any, string>;
  }
  public registerExternalService(
    service: IService,
  ): any {
    switch (service.protocol) {
      case "grpc":
        const grpcService = new GrpcService(service);
        this.externalServices.set(service, grpcService);
        return true;
      default:
        throw new Error(`Unsupported service protocol: ${service.protocol}`);
    }
  }

  public connectService(service: IService): any {
    const externalService = this.externalServices.get(service);
    if (!externalService) {
      throw new Error(`Service ${service.name} not found`);
    }
    return externalService.createClient();
  }

  public disconnectService(service: IService): any {
    const externalService = this.externalServices.get(service);
    if (!externalService) {
      throw new Error(`Service ${service.name} not found`);
    }
    return externalService.deleteClient();
  }

  public healthCheckAndUpdateService(service: IService): Promise<boolean> {
    const externalService = this.externalServices.get(service);
    if (!externalService) {
      throw new Error(`Service ${service.name} not found`);
    }
    return externalService.healthCheckAndUpdateClient();
  }

  public pingService(service: IService): Promise<boolean> {
    const externalService = this.externalServices.get(service);
    if (!externalService) {
      throw new Error(`Service ${service.name} not found`);
    }
    return externalService.ping();
  }

  public getClient(service: IService): any {
    const externalService = this.externalServices.get(service);
    if (!externalService) {
      throw new Error(`Service ${service.name} not found`);
    }
    return externalService.client;
  }

  public getAllClients(): any[] {
    return Array.from(this.externalServices.values()).map((externalService) => externalService.client);
  }

  public getAllServices(): IService[] {
    return Array.from(this.externalServices.keys());
  }

  public getAllExternalServices(): ExternalService[] {
    return Array.from(this.externalServices.values());
  }

  public createQueue(service: IService): any {
    switch (service.protocol) {
      case "grpc":
        const grpcService = this.externalServices.get(service);
        if (!grpcService) {
          throw new Error(`Service ${service.name} not found`);
        }
        return grpcService.createQueue();
      default:
        throw new Error(`Unsupported service protocol: ${service.protocol}`);
    }
  }

}