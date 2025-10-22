import express from 'express';
import { createBullBoard } from '@bull-board/api';
import { BullMQAdapter } from '@bull-board/api/bullMQAdapter';
import { ExpressAdapter } from '@bull-board/express';
import { ServiceQueueInfo } from '../queue/queue';

// Tạo Express adapter cho Bull Board
const serverAdapter = new ExpressAdapter();
serverAdapter.setBasePath(process.env.BULL_BOARD_PATH || '/admin/queues');

/**
 * Khởi tạo Bull Board với danh sách queues động
 */
export function initBullBoard(queues: ServiceQueueInfo[]) {
  const queueAdapters = queues.map(info => new BullMQAdapter(info.queue));

  createBullBoard({
    queues: queueAdapters,
    serverAdapter: serverAdapter,
  });

  console.log(`🎨 Bull Board initialized with ${queues.length} queue(s)`);
}

/**
 * Tạo Express app cho UI
 */
export function createBullBoardApp() {
  const app = express();
  app.use(process.env.BULL_BOARD_PATH || '/admin/queues', serverAdapter.getRouter());
  return app;
}

export { serverAdapter };
