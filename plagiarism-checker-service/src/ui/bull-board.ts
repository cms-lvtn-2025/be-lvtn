import express from 'express';
import path from 'path';
import { createBullBoard } from '@bull-board/api';
import { BullMQAdapter } from '@bull-board/api/bullMQAdapter';
import { ExpressAdapter } from '@bull-board/express';
import { ServiceQueueInfo } from '../queue/queue';
import adminRoutes from './admin-routes';

// Tạo Express adapter cho Bull Board
const serverAdapter = new ExpressAdapter();
serverAdapter.setBasePath(process.env.BULL_BOARD_PATH || '/admin/queues');

const buildQueueAdapters = (queues: ServiceQueueInfo[]) =>
  queues.map(info => new BullMQAdapter(info.queue));

let bullBoardInitialized = false;
let bullBoardController: ReturnType<typeof createBullBoard> | null = null;

/**
 * Khởi tạo hoặc đồng bộ Bull Board với danh sách queues động
 */
export function initBullBoard(queues: ServiceQueueInfo[]) {
  const queueAdapters = buildQueueAdapters(queues);

  if (!bullBoardInitialized) {
    bullBoardController = createBullBoard({
      queues: queueAdapters,
      serverAdapter: serverAdapter,
    });
    bullBoardInitialized = true;
    console.log(`🎨 Bull Board initialized with ${queues.length} queue(s)`);
  } else {
    bullBoardController?.setQueues(queueAdapters);
    console.log(`🔄 Bull Board updated with ${queues.length} queue(s)`);
  }
}

/**
 * Tạo Express app cho UI với Pug views và admin routes
 */
export function createBullBoardApp() {
  const app = express();

  // Setup Pug view engine
  app.set('view engine', 'pug');
  app.set('views', path.join(__dirname, 'views'));

  // JSON middleware cho API routes
  app.use(express.json());

  // Serve static files (CSS, JS)
  app.use('/public', express.static(path.join(__dirname, 'public')));

  // Mount admin routes (bao gồm cả SSR pages và API endpoints)
  app.use('/admin', adminRoutes);

  // Mount Bull Board UI (iframe source)
  app.use(process.env.BULL_BOARD_PATH || '/admin/queues', serverAdapter.getRouter());

  return app;
}

export { serverAdapter };
