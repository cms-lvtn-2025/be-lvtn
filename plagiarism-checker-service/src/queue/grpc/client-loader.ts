import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import { IService } from '../../database/models';
import fs from 'fs';
import path from 'path';

/**
 * Load gRPC client động từ service config
 */
export function loadGrpcClient(service: IService): any {
  try {
    if (!service.protoPath || !service.protoPackage) {
      throw new Error(`Service ${service.name} missing protoPath or protoPackage`);
    }

    console.log(`  📦 Loading proto: ${service.protoPath}`);

    // Load proto file với includeDirs để resolve import paths
    const protoDir = path.dirname(service.protoPath);
    const protoBaseDir = path.resolve('/home/thaily/code/heheheh_be/proto');
    const protoParentDir = path.resolve('/home/thaily/code/heheheh_be'); // Thêm parent dir

    const packageDefinition = protoLoader.loadSync(service.protoPath, {
      keepCase: true,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true,
      includeDirs: [protoDir, protoBaseDir, protoParentDir], // Thêm parent để resolve "proto/common/..."
    });

    const protoDescriptor = grpc.loadPackageDefinition(packageDefinition);

    // Get service constructor từ package path động (e.g., "file.FileService")
    const packagePath = service.protoPackage.split('.');
    let ServiceConstructor: any = protoDescriptor;

    for (const part of packagePath) {
      ServiceConstructor = ServiceConstructor[part];
      if (!ServiceConstructor) {
        throw new Error(`Cannot find package: ${service.protoPackage}`);
      }
    }

    // Tạo credentials (check TLS từ env)
    let credentials: grpc.ChannelCredentials;
    const useTLS = process.env.GRPC_TLS_ENABLED === 'true';

    if (useTLS) {
      try {
        const certPath = path.resolve(process.env.GRPC_CERT_PATH || '');
        const keyPath = path.resolve(process.env.GRPC_KEY_PATH || '');
        const caPath = path.resolve(process.env.GRPC_CA_PATH || '');

        const cert = fs.readFileSync(certPath);
        const key = fs.readFileSync(keyPath);
        const ca = fs.readFileSync(caPath);

        credentials = grpc.credentials.createSsl(ca, key, cert);
        console.log(`  🔒 TLS enabled`);
      } catch (error) {
        console.warn(`  ⚠️  TLS error, falling back to insecure`);
        credentials = grpc.credentials.createInsecure();
      }
    } else {
      credentials = grpc.credentials.createInsecure();
      console.log(`  🔓 Insecure connection`);
    }

    // Tạo client
    const address = `${service.url}:${service.port}`;
    const client = new ServiceConstructor(address, credentials);

    console.log(`  ✅ Client loaded for ${service.name} at ${address}`);

    return client;
  } catch (error) {
    console.error(`  ❌ Failed to load client for ${service.name}:`, error);
    throw error;
  }
}
