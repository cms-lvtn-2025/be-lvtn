/**
 * Type definitions for evaluateJob VM context
 *
 * IMPORTANT: Your code will be wrapped in an async function, so you can:
 * - Use `await` for async operations
 * - Use `return` to return values directly
 * - Or set `returnValue` variable
 *
 * Example:
 * ```typescript
 * const workflow = await findById(data.workflowId);
 * const result = await createJobWithChildren('QUEUE', 'EnJob', {}, []);
 * return result; // or: returnValue = result;
 * ```
 */

interface JobsOptions {
  [key: string]: any;
}

interface FlowOpts {
  [key: string]: any;
}

interface Children {
  serviceName: string;
  method: string;
  params: any;
  options?: JobsOptions;
  children?: Children[];
}

interface IWorkflowChildren {
  serviceName: string;
  method: string;
  params: any;
  options?: JobsOptions;
  children: IWorkflowChildren[];
}

interface IWorkflow {
  id?: string;
  name: string;
  parentServiceName: string;
  parentMethod: string;
  parentParams: any;
  children: IWorkflowChildren[];
  options?: JobsOptions;
  flowOpts?: FlowOpts;
}

declare global {
  /**
   * Return value variable (optional)
   * You can either set this variable OR use `return` statement
   */
  let returnValue: any;

  /**
   * Input data passed from job params
   */
  const data: any;

  /**
   * Console for logging
   */
  const console: {
    log(...args: any[]): void;
    error(...args: any[]): void;
    warn(...args: any[]): void;
    info(...args: any[]): void;
  };

  /**
   * Generate a UUID v4 string
   */
  function uuidv4(): string;

  /**
   * Find a document by ID (async)
   * @param id - Document ID
   * @example
   * const workflow = await findById('123abc');
   */
  function findById(id: any): Promise<any>;

  /**
   * Create a job with children jobs (async)
   * @param parentServiceName - Parent service name
   * @param parentMethod - Parent method name
   * @param parentParams - Parent job parameters
   * @param children - Array of child job configurations
   * @param options - Job options (optional)
   * @param FlowOpts - Flow options (optional)
   * @example
   * const job = await createJobWithChildren(
   *   'QUEUE',
   *   'EnJob',
   *   { data: 'test' },
   *   [{ serviceName: 'SERVICE', method: 'process', params: {} }]
   * );
   */
  function createJobWithChildren(
    parentServiceName: string,
    parentMethod: string,
    parentParams: any,
    children: Children[],
    options?: JobsOptions,
    FlowOpts?: FlowOpts
  ): Promise<any>;

  /**
   * Add data for workflow execution (async)
   * @param workFlow - Workflow object
   * @param data - Data to add
   */
  function addDataForWorkFlow(workFlow: IWorkflow, data: any): Promise<any>;

  /**
   * Date constructor
   */
  const Date: DateConstructor;
}

export {};
