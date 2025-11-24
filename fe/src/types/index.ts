export interface User {
  id: string;
  email: string;
  role: 'admin' | 'user';
  twoFactorEnabled: boolean;
  lastLogin?: string;
  tokenExpiresAt?: string;
}

export interface LoginResponse {
  success: boolean;
  message: string;
  data: {
    tempToken: string;
    userId: string;
    email: string;
    twoFactorEnabled: boolean;
    requiresVerification: boolean;
  };
}

export interface Setup2FAResponse {
  success: boolean;
  message: string;
  data: {
    secret: string;
    qrCodeUrl: string;
    manualEntryKey: string;
  };
}

export interface VerifyResponse {
  success: boolean;
  message: string;
  data: {
    accessToken: string;
    expiresIn: string;
    expiresAt: string;
    user: User;
  };
}

export interface Service {
  _id: string;
  name: string;
  url: string;
  port: number;
  protocol: 'grpc' | 'http' | 'https';
  protoPath?: string;
  protoPackage?: string;
  enabled: boolean;
  healthy: boolean;
  lastHealthCheck?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Workflow {
  _id: string;
  name: string;
  parentServiceName: string;
  parentMethod: string;
  parentParams?: any;
  children?: WorkflowChild[];
  options?: any;
  createdAt: string;
  updatedAt: string;
  cronJob?: CronJob;
  hasCronJob?: boolean;
}

export interface WorkflowChild {
  serviceName: string;
  method: string;
  params?: any;
  options?: any;
  children?: WorkflowChild[];
}

export interface CronJob {
  _id: string;
  schedule: string; // Backend uses "schedule" not "cronExpression"
  WL_id: string;
  enabled: boolean;
  lastRun?: string;
  nextRun?: string;
  idJobCureent?: string;
  createdAt: string;
  updatedAt: string;
}

export interface MinioConfig {
  _id: string;
  name: string;
  endPoint: string;
  port: number;
  useSSL: boolean;
  accessKey: string;
  secretKey: string;
  bucketName: string;
  enabled: boolean;
  connectionStatus?: {
    connected: boolean;
    lastCheck?: string;
    error?: string;
  };
  createdAt: string;
  updatedAt: string;
}

export interface Queue {
  name: string;
  type: 'dynamic' | 'static';
  jobCounts: {
    active: number;
    waiting: number;
    completed: number;
    failed: number;
    delayed?: number;
  };
}

export interface Job {
  id: string;
  name: string;
  data: any;
  progress: number;
  returnvalue: any;
  failedReason?: string;
  state: string;
  timestamp: number;
  processedOn?: number;
  finishedOn?: number;
}

export interface ApiResponse<T> {
  success: boolean;
  message?: string;
  data?: T;
  count?: number;
  total?: number;
  errors?: Array<{ field: string; message: string }>;
}
