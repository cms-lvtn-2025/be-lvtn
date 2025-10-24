import express, { Request, Response } from 'express';
import { ServiceModel, MinioConfigModel, WorkflowModel } from '../database/models';
import { MinioService } from '../queue/minio';
import { serviceQueueManager } from '../queue/queue';

const router = express.Router();

/**
 * GET /admin - BullMQ Dashboard (main page)
 */
router.get('/', (req: Request, res: Response) => {
  res.render('bullmq', {
    title: 'BullMQ Dashboard',
    currentTab: 'bullmq',
    bullmqPath: process.env.BULL_BOARD_PATH || '/admin/queues'
  });
});

/**
 * GET /admin/workflow - Workflow Manager Page (SSR)
 */
router.get('/workflow', async (req: Request, res: Response) => {
  try {
    const workflows = await WorkflowModel.find({}).sort({ createdAt: -1 }).lean();

    res.render('workflow', {
      title: 'Workflow Manager',
      currentTab: 'workflow',
      workflows: workflows
    });
  } catch (error) {
    console.error('Error fetching workflows:', error);
    res.render('workflow', {
      title: 'Workflow Manager',
      currentTab: 'workflow',
      workflows: [],
      error: 'Failed to load workflows'
    });
  }
});

/**
 * GET /admin/workflow/editor/:id? - Workflow Editor Page (SSR)
 */
router.get('/workflow/editor/:id?', async (req: Request, res: Response) => {
  try {
    let workflow = null;

    if (req.params.id) {
      workflow = await WorkflowModel.findById(req.params.id).lean();
    }

    // Lấy danh sách services từ serviceQueueManager
    const allQueues = serviceQueueManager.getAllQueues();

    // Phân loại services theo type
    const servicesList = {
      fixed: [
        { name: 'QUEUE', type: 'QUEUE', methods: ['EnJob', 'evaluateJob'] },
        { name: 'MONGODB_WORKFLOW', type: 'WORKFLOW', methods: ['findById', 'find', 'findOne', 'create', 'update', 'delete'] }
      ],
      static: allQueues
        .filter(q => q.type === 'static' && q.serviceName !== 'QUEUE' && q.serviceName !== 'MONGODB_WORKFLOW')
        .map(q => ({
          name: q.serviceName,
          type: 'STATIC',
          methods: q.serviceName.startsWith('MINIO_SERVICE')
            ? ['uploadBuffer', 'getFile', 'deleteFile', 'listFiles', 'generateTemplate1PDF']
            : []
        })),
      dynamic: allQueues
        .filter(q => q.type === 'dynamic')
        .map(q => ({
          name: q.serviceName,
          type: 'DYNAMIC',
          service: q.service
        }))
    };

    res.render('workflow-editor', {
      title: 'Workflow Editor',
      currentTab: 'workflow',
      workflow: workflow,
      servicesList: servicesList
    });
  } catch (error) {
    console.error('Error fetching workflow:', error);
    res.render('workflow-editor', {
      title: 'Workflow Editor',
      currentTab: 'workflow',
      workflow: null,
      servicesList: { fixed: [], static: [], dynamic: [] },
      error: 'Failed to load workflow'
    });
  }
});

/**
 * GET /admin/services - Manager Service Page (SSR)
 */
router.get('/services', async (req: Request, res: Response) => {
  try {
    const services = await ServiceModel.find({}).sort({ serviceName: 1 }).lean();

    res.render('services', {
      title: 'Manager Service',
      currentTab: 'services',
      services: services
    });
  } catch (error) {
    console.error('Error fetching services:', error);
    res.render('services', {
      title: 'Manager Service',
      currentTab: 'services',
      services: [],
      error: 'Failed to load services'
    });
  }
});

/**
 * GET /admin/minio - Manager MinIO Page (SSR)
 */
router.get('/minio', async (req: Request, res: Response) => {
  try {
    // Lấy data từ MongoDB (đã có connectionStatus được update từ MinioService)
    const configs = await MinioConfigModel.find({}).sort({ name: 1 }).lean();

    res.render('minio', {
      title: 'Manager MinIO',
      currentTab: 'minio',
      minioConfigs: configs
    });
  } catch (error) {
    console.error('Error fetching MinIO configs:', error);
    res.render('minio', {
      title: 'Manager MinIO',
      currentTab: 'minio',
      minioConfigs: [],
      error: 'Failed to load MinIO configs'
    });
  }
});

export default router;
