import * as grpc from '@grpc/grpc-js';
import { IService, ServiceModel } from '../../database/models';
import { loadGrpcClient } from './client-loader';

/**
 * Ping gRPC service để kiểm tra health
 */
export async function pingGrpcService(service: IService): Promise<boolean> {
  try {
    console.log(`  🏓 Pinging ${service.name}...`);

    // Load client
    const client = loadGrpcClient(service);

    // Kiểm tra connection state
    const channel = client.getChannel();
    const state = channel.getConnectivityState(true); // true = try to connect

    // Wait for connection (timeout 5s)
    const timeout = service.metadata?.timeout || 5000;
    const connected = await waitForConnection(channel, timeout);

    if (connected) {
      console.log(`  ✅ ${service.name} is healthy`);
      return true;
    } else {
      console.log(`  ❌ ${service.name} connection timeout`);
      return false;
    }
  } catch (error: any) {
    console.error(`  ❌ ${service.name} health check failed:`, error.message);
    return false;
  }
}

/**
 * Wait for gRPC channel to connect
 */
function waitForConnection(
  channel: grpc.Channel,
  timeoutMs: number
): Promise<boolean> {
  return new Promise((resolve) => {
    const deadline = Date.now() + timeoutMs;

    const checkState = () => {
      const state = channel.getConnectivityState(false);

      // READY = connected
      if (state === grpc.connectivityState.READY) {
        resolve(true);
        return;
      }

      // SHUTDOWN or TRANSIENT_FAILURE = failed
      if (
        state === grpc.connectivityState.SHUTDOWN ||
        state === grpc.connectivityState.TRANSIENT_FAILURE
      ) {
        if (Date.now() >= deadline) {
          resolve(false);
          return;
        }
      }

      // Continue waiting
      if (Date.now() < deadline) {
        channel.watchConnectivityState(state, deadline, (error) => {
          if (error) {
            resolve(false);
          } else {
            checkState();
          }
        });
      } else {
        resolve(false);
      }
    };

    checkState();
  });
}

/**
 * Health check một service và update database
 */
export async function healthCheckAndUpdate(service: IService): Promise<boolean> {
  const isHealthy = await pingGrpcService(service);

  // Update database
  await ServiceModel.updateOne(
    { _id: service._id },
    {
      $set: {
        healthy: isHealthy,
        lastHealthCheck: new Date(),
      },
    }
  );

  return isHealthy;
}

/**
 * Health check tất cả services và update database
 */
export async function healthCheckAllServices(
  services: IService[]
): Promise<IService[]> {
  console.log('\n🏥 Running health checks...');

  const results = await Promise.all(
    services.map(async (service) => {
      const isHealthy = await healthCheckAndUpdate(service);
      return { ...service.toObject(), healthy: isHealthy };
    })
  );

  const healthyCount = results.filter((s) => s.healthy).length;
  console.log(`\n✅ Health check completed: ${healthyCount}/${results.length} services healthy`);

  return results;
}
