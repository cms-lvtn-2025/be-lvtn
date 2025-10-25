import { CronJobModel } from "../database/models";
import { cronJobService } from "./cronjob-service";
import { serviceQueueManager } from "./queue";

/**
 * Khởi tạo lại tất cả CronJobs khi server start
 * - Xóa tất cả repeatable jobs cũ trong BullMQ
 * - Tạo lại từ database
 */
export async function initializeCronJobs(): Promise<void> {
  try {
    console.log("\n🔄 Initializing CronJobs...");

    // 1. Lấy QUEUE service
    const queueInfo = serviceQueueManager.getServiceQueue("QUEUE");
    if (!queueInfo) {
      console.warn("⚠️  QUEUE service not found, skipping CronJob initialization");
      return;
    }

    // 2. Xóa TẤT CẢ repeatable jobs cũ trong BullMQ
    console.log("🗑️  Cleaning old repeatable jobs from BullMQ...");
    const repeatableJobs = await queueInfo.queue.getRepeatableJobs();

    for (const job of repeatableJobs) {
      try {
        await queueInfo.queue.removeRepeatableByKey(job.key);
        console.log(`   ✓ Removed old job: ${job.id || job.key}`);
      } catch (error: any) {
        console.warn(`   ⚠️  Could not remove job ${job.key}: ${error.message}`);
      }
    }
    console.log(`✅ Cleaned ${repeatableJobs.length} old repeatable jobs`);

    // 3. Lấy tất cả CronJobs từ database
    const cronJobs = await CronJobModel.find({ enabled: true });

    if (cronJobs.length === 0) {
      console.log("ℹ️  No enabled CronJobs found in database");
      return;
    }

    console.log(`📋 Found ${cronJobs.length} enabled CronJob(s) in database`);

    // 4. Tạo lại tất cả CronJobs trong BullMQ
    let successCount = 0;
    let errorCount = 0;

    for (const cronJob of cronJobs) {
      try {
        const { v4: uuidv4 } = require("uuid");
        const newJobId = uuidv4();

        // Tạo repeatable job trực tiếp trong BullMQ (không qua createCronJob để tránh duplicate check)
        await queueInfo.queue.add(
          "grpc-call",
          {
            method: "evaluateJob",
            params: {
              code: `
                const workflow = await findById("${cronJob.WL_id}");
                if (!workflow) {
                  throw new Error("Workflow not found with ID: ${cronJob.WL_id}");
                }
                await createJobWithChildren(
                  workflow.parentServiceName,
                  workflow.parentMethod,
                  workflow.parentParams,
                  workflow.children,
                  workflow.options
                );
                return "Successfully created child jobs for workflow ID: ${cronJob.WL_id}";
              `,
              returnValue: {
                workflowId: cronJob.WL_id,
              },
            },
            metadata: {
              serviceName: "QUEUE",
              workflowId: cronJob.WL_id,
            },
          },
          {
            jobId: newJobId,
            repeat: {
              pattern: cronJob.schedule,
            },
          }
        );

        // Cập nhật jobId mới vào database
        cronJob.idJobCureent = newJobId;
        await cronJob.save();

        console.log(`   ✓ Recreated CronJob for workflow ${cronJob.WL_id}: ${cronJob.schedule} (${newJobId})`);
        successCount++;
      } catch (error: any) {
        console.error(`   ❌ Failed to recreate CronJob for workflow ${cronJob.WL_id}: ${error.message}`);
        errorCount++;
      }
    }

    console.log(`\n✅ CronJobs initialized: ${successCount} success, ${errorCount} errors\n`);
  } catch (error: any) {
    console.error("❌ Failed to initialize CronJobs:", error.message);
    throw error;
  }
}

/**
 * Dọn dẹp tất cả repeatable jobs khi shutdown
 */
export async function cleanupCronJobs(): Promise<void> {
  try {
    console.log("\n🧹 Cleaning up CronJobs...");

    const queueInfo = serviceQueueManager.getServiceQueue("QUEUE");
    if (!queueInfo) {
      return;
    }

    const repeatableJobs = await queueInfo.queue.getRepeatableJobs();

    for (const job of repeatableJobs) {
      await queueInfo.queue.removeRepeatableByKey(job.key);
    }

    console.log(`✅ Cleaned up ${repeatableJobs.length} repeatable jobs`);
  } catch (error: any) {
    console.error("❌ Failed to cleanup CronJobs:", error.message);
  }
}
