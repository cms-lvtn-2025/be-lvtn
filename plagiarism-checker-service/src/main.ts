import dotenv from 'dotenv';
import DatabaseConnection from './database/connection';
import { ServiceModel } from './database/models';
import { healthCheckAllServices } from './queue/grpc';
import { initBullBoard, createBullBoardApp } from './ui/bull-board';
import { queueService, serviceQueueManager } from './queue/queue';
import { v4 as uuidv4 } from 'uuid';

// Load environment variables
dotenv.config();

/**
 * Main application entry point
 */
async function main() {
  try {
    console.log('🚀 Starting Plagiarism Checker Service...');
    console.log(`Environment: ${process.env.NODE_ENV || 'development'}`);

    // Connect to MongoDB
    await DatabaseConnection.connect();

    // Load tất cả enabled services từ database
    const allServices = await ServiceModel.find({ enabled: true });
    console.log(`\n📊 Found ${allServices.length} enabled service(s)`);

    // Health check tất cả services và update database
    const healthyServices = await healthCheckAllServices(allServices);

    // Chỉ tạo queue cho services healthy
    const servicesToQueue = healthyServices.filter(s => s.healthy);
    console.log(`\n🔧 Creating queues for ${servicesToQueue.length} healthy service(s)...`);

    const queues = servicesToQueue.map(service => {
      return serviceQueueManager.createServiceQueue(service as any);
    });

    // Khởi tạo Bull Board UI với tất cả queues
    initBullBoard(queues);

    // Start Bull Board UI
    const bullBoardApp = createBullBoardApp();
    const port = parseInt(process.env.BULL_BOARD_PORT || '3000');
    bullBoardApp.listen(port, () => {
      console.log(`\n🎨 Bull Board UI: http://localhost:${port}${process.env.BULL_BOARD_PATH || '/admin/queues'}`);
    });

    // Example: Thêm test job vào FILE_SERVICE queue
    console.log('\n📝 Adding test job to FILE_SERVICE...');
    const { v4: uuidv4 } = require('uuid');
    const childJobId = uuidv4();
    const nameService = "FILE_SERVICE";
    const queueName = `${nameService.toLowerCase()}-queue`;

    console.log('Child Job ID:', childJobId);

    await queueService.createJobWithChildren(nameService, 'UpdateFile', {
      // Reference kết quả từ child job (custom syntax với @bull:)
      id: `@bull:${queueName}:${childJobId}.file.id`,
      status: "REJECTED",
    }, [
      {
        id: childJobId, // Job ID của child
        serviceName: nameService,
        method: 'GetFile',
        params: { id: '8f34e987-b48e-4fdf-9fff-47d932a80c4a' }
      }
    ])

    console.log('\n✨ Application is running...');
    console.log(`   - ${servicesToQueue.length} service queue(s) active`);
    console.log(`   - Bull Board UI running on port ${port}`);
    console.log('\nPress Ctrl+C to exit\n');

  } catch (error) {
    console.error('❌ Failed to start application:', error);
    process.exit(1);
  }
}

// Handle graceful shutdown
process.on('SIGINT', async () => {
  console.log('\n\n🛑 Shutting down gracefully...');
  await queueService.close();
  await serviceQueueManager.closeAll();
  await DatabaseConnection.disconnect();
  process.exit(0);
});

process.on('SIGTERM', async () => {
  console.log('\n\n🛑 Shutting down gracefully...');
  await queueService.close();
  await serviceQueueManager.closeAll();
  await DatabaseConnection.disconnect();
  process.exit(0);
});

// Start the application
main().catch((error) => {
  console.error('Fatal error:', error);
  process.exit(1);
});
