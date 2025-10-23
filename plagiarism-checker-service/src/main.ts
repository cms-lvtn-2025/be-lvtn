import dotenv from "dotenv";
import DatabaseConnection from "./database/connection";
import { ServiceModel, WorkflowModel } from "./database/models";
import { healthCheckAllServices } from "./queue/grpc";
import { initBullBoard, createBullBoardApp } from "./ui/bull-board";
import { queueService, serviceQueueManager } from "./queue/queue";
import { v4 as uuidv4 } from "uuid";
import cron from 'node-cron';

// Load environment variables
dotenv.config();

/**
 * Main application entry point
 */
async function main() {
  try {
    console.log("🚀 Starting Plagiarism Checker Service...");
    console.log(`Environment: ${process.env.NODE_ENV || "development"}`);

    // Connect to MongoDB
    await DatabaseConnection.connect();

    // Load tất cả enabled services từ database
    const allServices = await ServiceModel.find({ enabled: true });
    console.log(`\n📊 Found ${allServices.length} enabled service(s)`);

    // Health check tất cả services và update database
    const healthyServices = await healthCheckAllServices(allServices);

    // Chỉ tạo queue cho services healthy
    const servicesToQueue = healthyServices.filter((s) => s.healthy);
    console.log(
      `\n🔧 Creating queues for ${servicesToQueue.length} healthy service(s)...`
    );

    const queues = servicesToQueue.map((service) => {
      return serviceQueueManager.createServiceQueue(service as any);
    });

    const queuesStatic = serviceQueueManager.registerStaticQueue(
      "QUEUE",
      queueService,
      {
        concurrency: 2,
        attempts: 3,
      }
    );
    queues.push(queuesStatic);

    const queueWorkflow =
      serviceQueueManager.createServiceWorkflowQueue(WorkflowModel);
    queues.push(queueWorkflow);
    // Khởi tạo Bull Board UI với tất cả queues
    initBullBoard(queues);

    // Start Bull Board UI
    const bullBoardApp = createBullBoardApp();
    const port = parseInt(process.env.BULL_BOARD_PORT || "3000");
    bullBoardApp.listen(port, () => {
      console.log(
        `\n🎨 Bull Board UI: http://localhost:${port}${
          process.env.BULL_BOARD_PATH || "/admin/queues"
        }`
      );
    });
    // const WorkflowData = await WorkflowModel.findById("68f9d792133085ee4f6900b4") ;
    // if (WorkflowData) {
    //   await queueService.createJobWithChildren(WorkflowData.parentServiceName, WorkflowData.parentMethod, WorkflowData.parentParams, WorkflowData.children, WorkflowData.options);
    // }
    // Example: Thêm test job vào FILE_SERVICE queue
    console.log("\n📝 Adding test job to FILE_SERVICE...");
    const { v4: uuidv4 } = require("uuid");
    const childJobId2 = uuidv4();
    const childJobId1 = uuidv4();

    const parentJobId = uuidv4();
    const nameService = "FILE_SERVICE";

    console.log("Child Job ID:", childJobId2);

    cron.schedule("* * * * *", async () => {
      await queueService.createJobWithChildren(
      "QUEUE",
      "EnJob",
      {
        // Reference kết quả từ child job (custom syntax với @bull:)
        id: `@bull:${nameService}:${childJobId2}.file.id`,
        status: "APPROVED",
      },
      [
        {
          serviceName: "QUEUE",
          method: "evaluateJob",
          params: {
            code: `
                console.log("xxxxxxxxxxxxxxxxxxxxx", returnValue)
                await createJobWithChildren(returnValue.parentServiceName, returnValue.parentMethod, returnValue.parentParams, returnValue.children, returnValue.options);
                return "Successfully created child jobs for each file."
              `,
            returnValue: `@__id__0:`,
          },
          children: [
            {
              serviceName: "MONGODB_WORKFLOW",
              method: "findById",
              params: "68f9d792133085ee4f6900b4",
            },
          ],
        },
      ],
      {
        repeat: {
          pattern: "* * * * *", // Chạy mỗi phút
        },
      }
    );
    })

    

    console.log("\n✨ Application is running...");
    console.log(`   - ${servicesToQueue.length} service queue(s) active`);
    console.log(`   - Bull Board UI running on port ${port}`);
    console.log("\nPress Ctrl+C to exit\n");
  } catch (error) {
    console.error("❌ Failed to start application:", error);
    process.exit(1);
  }
}

// Handle graceful shutdown
process.on("SIGINT", async () => {
  console.log("\n\n🛑 Shutting down gracefully...");
  await queueService.close();
  await serviceQueueManager.closeAll();
  await DatabaseConnection.disconnect();
  process.exit(0);
});

process.on("SIGTERM", async () => {
  console.log("\n\n🛑 Shutting down gracefully...");
  await queueService.close();
  await serviceQueueManager.closeAll();
  await DatabaseConnection.disconnect();
  process.exit(0);
});

// Start the application
main().catch((error) => {
  console.error("Fatal error:", error);
  process.exit(1);
});
