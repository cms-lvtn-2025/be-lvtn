import React, { useCallback, useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import ReactFlow, {
  Controls,
  Background,
  useNodesState,
  useEdgesState,
  addEdge,
  MarkerType,
  BackgroundVariant,
  reconnectEdge,
  useReactFlow,
  ReactFlowProvider,
  type Node,
  type Edge,
  type Connection,
} from 'reactflow';
import 'reactflow/dist/style.css';
import dagre from 'dagre';
import Editor from '@monaco-editor/react';
import WorkflowNode from './WorkflowNode';
import NodePalette from './NodePalette';
import { workflowsApi } from '../api/workflows';
import type { Workflow } from '../types';

interface WorkflowEditorProps {
  workflowId?: string;
  onSave?: () => void;
  onCancel?: () => void;
}

const nodeTypes = {
  workflow: WorkflowNode,
};

// Auto-layout using dagre
const getLayoutedElements = (nodes: Node[], edges: Edge[], direction = 'LR') => {
  const dagreGraph = new dagre.graphlib.Graph();
  dagreGraph.setDefaultEdgeLabel(() => ({}));

  const nodeWidth = 220;
  const nodeHeight = 150;

  const isHorizontal = direction === 'LR';
  dagreGraph.setGraph({
    rankdir: direction,
    nodesep: 80,  // Horizontal spacing between nodes
    ranksep: 150, // Vertical spacing between ranks
    edgesep: 50,  // Spacing between edges
  });

  nodes.forEach((node) => {
    dagreGraph.setNode(node.id, { width: nodeWidth, height: nodeHeight });
  });

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target);
  });

  dagre.layout(dagreGraph);

  nodes.forEach((node) => {
    const nodeWithPosition = dagreGraph.node(node.id);
    node.targetPosition = isHorizontal ? 'left' : 'top';
    node.sourcePosition = isHorizontal ? 'right' : 'bottom';

    // We are shifting the dagre node position (anchor=center center) to the top left
    // so it matches the React Flow node anchor point (top left).
    node.position = {
      x: nodeWithPosition.x - nodeWidth / 2,
      y: nodeWithPosition.y - nodeHeight / 2,
    };

    return node;
  });

  return { nodes, edges };
};

// Inner component that uses ReactFlow hooks
const WorkflowEditorInner: React.FC<WorkflowEditorProps> = ({ workflowId: propWorkflowId, onSave, onCancel }) => {
  const { id: paramWorkflowId } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const workflowId = propWorkflowId || paramWorkflowId;

  const [nodes, setNodes, onNodesChange] = useNodesState([]);
  const [edges, setEdges, onEdgesChange] = useEdgesState([]);
  const [workflowName, setWorkflowName] = useState('');
  const [saving, setSaving] = useState(false);
  const [availableServices, setAvailableServices] = useState<any>(null);
  const [selectedNode, setSelectedNode] = useState<Node | null>(null);
  const [editedNodeData, setEditedNodeData] = useState<any>(null);
  const [paramsText, setParamsText] = useState<string>('');
  const [optionsText, setOptionsText] = useState<string>('');
  const { screenToFlowPosition } = useReactFlow();

  const handleNavigateBack = () => {
    if (onCancel) {
      onCancel();
    } else {
      navigate('/workflows');
    }
  };

  const handleSaveSuccess = () => {
    if (onSave) {
      onSave();
    } else {
      navigate('/workflows');
    }
  };

  useEffect(() => {
    loadAvailableServices();
    if (workflowId) {
      loadWorkflow();
    } else {
      // Create default parent node for new workflow
      createDefaultParentNode();
    }
  }, [workflowId]);

  const loadAvailableServices = async () => {
    try {
      const response = await workflowsApi.getAvailableServices();
      setAvailableServices(response.data);
    } catch (error) {
      console.error('Failed to load services:', error);
    }
  };

  const loadWorkflow = async () => {
    if (!workflowId) return;

    try {
      const response = await workflowsApi.getById(workflowId);
      const workflow = response.data;

      setWorkflowName(workflow.name);

      // Convert workflow data to ReactFlow nodes and edges
      const flowNodes = convertWorkflowToNodes(workflow);
      const flowEdges = convertWorkflowToEdges(workflow);

      // Automatically apply layout to position children before parent
      const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
        flowNodes,
        flowEdges,
        'LR' // Left to Right: children on left, parent on right (matches execution order)
      );

      setNodes(layoutedNodes);
      setEdges(layoutedEdges);
    } catch (error) {
      console.error('Failed to load workflow:', error);
      alert('Failed to load workflow');
    }
  };

  // Update nodes with delete callbacks after they're set
  useEffect(() => {
    setNodes((nds) =>
      nds.map((node) => ({
        ...node,
        data: {
          ...node.data,
          onDelete: node.data.isParent ? undefined : () => handleDeleteNode(node.id),
        },
      }))
    );
  }, [workflowId]); // Only run when workflow is loaded

  const createDefaultParentNode = () => {
    const parentNode: Node = {
      id: 'parent',
      type: 'workflow',
      position: { x: 400, y: 100 },
      data: {
        label: 'Parent Job',
        serviceName: 'QUEUE',
        method: 'EnJob',
        params: {},
        options: {},
        isParent: true,
      },
    };

    setNodes([parentNode]);
  };

  const convertWorkflowToNodes = (workflow: Workflow): Node[] => {
    const nodes: Node[] = [];
    let nodeId = 1;

    // Parent node
    nodes.push({
      id: 'parent',
      type: 'workflow',
      position: { x: 400, y: 100 },
      data: {
        label: 'Parent Job',
        serviceName: workflow.parentServiceName,
        method: workflow.parentMethod,
        params: workflow.parentParams,
        options: workflow.options || {},
        isParent: true,
      },
    });

    // Children nodes
    const addChildren = (children: any[], parentId: string, startX: number, startY: number) => {
      children?.forEach((child, index) => {
        const currentId = `node-${nodeId++}`;
        const y = startY + (index * 150);

        nodes.push({
          id: currentId,
          type: 'workflow',
          position: { x: startX, y },
          data: {
            label: child.serviceName,
            serviceName: child.serviceName,
            method: child.method,
            params: child.params,
            options: child.options || {},
            isParent: false,
          },
        });

        // Recursively add nested children
        if (child.children && child.children.length > 0) {
          addChildren(child.children, currentId, startX - 200, y);
        }
      });
    };

    if (workflow.children) {
      addChildren(workflow.children, 'parent', 200, 100);
    }

    return nodes;
  };

  const convertWorkflowToEdges = (workflow: Workflow): Edge[] => {
    const edges: Edge[] = [];
    let nodeId = 1;

    const addChildEdges = (children: any[], parentId: string) => {
      children?.forEach((child) => {
        const currentId = `node-${nodeId++}`;

        edges.push({
          id: `e-${parentId}-${currentId}`,
          source: currentId,
          target: parentId,
          type: 'smoothstep',
          animated: true,
          style: { stroke: '#3498db', strokeWidth: 3 },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: '#3498db',
          },
        });

        if (child.children && child.children.length > 0) {
          addChildEdges(child.children, currentId);
        }
      });
    };

    if (workflow.children) {
      addChildEdges(workflow.children, 'parent');
    }

    return edges;
  };

  const onConnect = useCallback(
    (params: Connection) => {
      const newEdge = {
        ...params,
        type: 'smoothstep',
        animated: true,
        style: { stroke: '#3498db', strokeWidth: 3 },
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: '#3498db',
        },
      };
      setEdges((eds) => addEdge(newEdge, eds));
    },
    [setEdges]
  );

  const onEdgeClick = useCallback((event: React.MouseEvent, edge: Edge) => {
    event.stopPropagation();
    if (confirm('Delete this connection?')) {
      setEdges((eds) => eds.filter((e) => e.id !== edge.id));
    }
  }, [setEdges]);

  // Handle edge reconnection (drag edge endpoint to another node)
  const edgeReconnectSuccessful = React.useRef(true);

  const onReconnectStart = useCallback(() => {
    edgeReconnectSuccessful.current = false;
  }, []);

  const onReconnect = useCallback((oldEdge: Edge, newConnection: Connection) => {
    edgeReconnectSuccessful.current = true;
    setEdges((els) => reconnectEdge(oldEdge, newConnection, els));
  }, [setEdges]);

  const onReconnectEnd = useCallback((_: any, edge: Edge) => {
    if (!edgeReconnectSuccessful.current) {
      setEdges((eds) => eds.filter((e) => e.id !== edge.id));
    }
    edgeReconnectSuccessful.current = true;
  }, [setEdges]);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();

      const type = event.dataTransfer.getData('application/reactflow');
      const serviceName = event.dataTransfer.getData('serviceName');
      const method = event.dataTransfer.getData('method');

      if (!type || !serviceName) {
        return;
      }

      // Convert screen position to flow position (accounts for zoom and pan)
      const position = screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      });

      const newNodeId = `node-${Date.now()}`;

      const newNode: Node = {
        id: newNodeId,
        type: 'workflow',
        position,
        data: {
          label: serviceName,
          serviceName,
          method: method || 'EnJob',
          params: {},
          options: {},
          isParent: false,
          onDelete: () => handleDeleteNode(newNodeId),
        },
      };

      setNodes((nds) => nds.concat(newNode));
    },
    [screenToFlowPosition]
  );

  const handleDeleteNode = (nodeId: string) => {
    if (confirm('Delete this node and its connections?')) {
      setNodes((nds) => nds.filter((n) => n.id !== nodeId));
      setEdges((eds) => eds.filter((e) => e.source !== nodeId && e.target !== nodeId));

      // Close edit panel if this node was being edited
      if (selectedNode?.id === nodeId) {
        setSelectedNode(null);
        setEditedNodeData(null);
      }
    }
  };

  const handleClearCanvas = () => {
    if (confirm('Clear all nodes except parent node?')) {
      setNodes((nds) => nds.filter((n) => n.data.isParent));
      setEdges([]);
      setSelectedNode(null);
      setEditedNodeData(null);
    }
  };

  const handleAutoLayout = useCallback(() => {
    const { nodes: layoutedNodes, edges: layoutedEdges } = getLayoutedElements(
      nodes,
      edges,
      'LR' // Left to Right
    );

    setNodes([...layoutedNodes]);
    setEdges([...layoutedEdges]);

    // Fit view after layout
    setTimeout(() => {
      window.requestAnimationFrame(() => {
        const reactFlowInstance = (window as any).reactFlowInstance;
        if (reactFlowInstance) {
          reactFlowInstance.fitView({ padding: 0.2 });
        }
      });
    }, 0);
  }, [nodes, edges, setNodes, setEdges]);

  const handleSaveWorkflow = async () => {
    if (!workflowName.trim()) {
      alert('Please enter a workflow name');
      return;
    }

    setSaving(true);

    try {
      // Convert ReactFlow nodes/edges back to workflow structure
      const parentNode = nodes.find(n => n.data.isParent);
      if (!parentNode) {
        alert('Parent node is required');
        return;
      }

      const workflowData = {
        name: workflowName,
        parentServiceName: parentNode.data.serviceName,
        parentMethod: parentNode.data.method,
        parentParams: parentNode.data.params || {},
        options: parentNode.data.options || {},
        children: buildChildrenStructure(),
      };

      if (workflowId) {
        await workflowsApi.update(workflowId, workflowData);
        alert('Workflow updated successfully!');
      } else {
        await workflowsApi.create(workflowData);
        alert('Workflow created successfully!');
      }

      handleSaveSuccess();
    } catch (error: any) {
      console.error('Failed to save workflow:', error);
      alert(error.response?.data?.message || 'Failed to save workflow');
    } finally {
      setSaving(false);
    }
  };

  const buildChildrenStructure = (): any[] => {
    const children: any[] = [];
    const parentNode = nodes.find(n => n.data.isParent);

    if (!parentNode) return [];

    // Find all nodes connected to parent
    const childEdges = edges.filter(e => e.target === parentNode.id);

    childEdges.forEach(edge => {
      const childNode = nodes.find(n => n.id === edge.source);
      if (childNode && !childNode.data.isParent) {
        children.push({
          serviceName: childNode.data.serviceName,
          method: childNode.data.method,
          params: childNode.data.params || {},
          options: childNode.data.options || {},
          children: buildChildrenForNode(childNode.id),
        });
      }
    });

    return children;
  };

  const buildChildrenForNode = (nodeId: string): any[] => {
    const children: any[] = [];
    const childEdges = edges.filter(e => e.target === nodeId);

    childEdges.forEach(edge => {
      const childNode = nodes.find(n => n.id === edge.source);
      if (childNode) {
        children.push({
          serviceName: childNode.data.serviceName,
          method: childNode.data.method,
          params: childNode.data.params || {},
          options: childNode.data.options || {},
          children: buildChildrenForNode(childNode.id),
        });
      }
    });

    return children;
  };

  const handleNodeClick = useCallback((event: React.MouseEvent, node: Node) => {
    setSelectedNode(node);
    setEditedNodeData({...node.data});

    // Initialize text states for textareas
    if (node.data.method === 'evaluateJob') {
      const { code, ...paramsWithoutCode } = node.data.params || {};
      setParamsText(JSON.stringify(paramsWithoutCode, null, 2));
    } else {
      setParamsText(JSON.stringify(node.data.params || {}, null, 2));
    }
    setOptionsText(JSON.stringify(node.data.options || {}, null, 2));
  }, []);

  const handleUpdateNodeData = () => {
    if (!selectedNode) return;

    // Validate evaluateJob requires code
    if (editedNodeData.method === 'evaluateJob') {
      if (!editedNodeData.params?.code || !editedNodeData.params.code.trim()) {
        alert('Code is required for evaluateJob method!\n\nPlease enter JavaScript/TypeScript code in the "Code" textarea.');
        return;
      }
    }

    // Parse params from text
    let parsedParams;
    try {
      parsedParams = JSON.parse(paramsText);
      // Preserve code if it exists
      if (editedNodeData.method === 'evaluateJob' && editedNodeData.params?.code) {
        parsedParams.code = editedNodeData.params.code;
      }
    } catch (err) {
      alert('Invalid JSON in Parameters field. Please check the syntax.');
      return;
    }

    // Parse options from text
    let parsedOptions;
    try {
      parsedOptions = JSON.parse(optionsText);
    } catch (err) {
      alert('Invalid JSON in Options field. Please check the syntax.');
      return;
    }

    // Update editedNodeData with parsed values
    const updatedData = {
      ...editedNodeData,
      params: parsedParams,
      options: parsedOptions,
    };

    setNodes((nds) =>
      nds.map((node) =>
        node.id === selectedNode.id
          ? { ...node, data: updatedData }
          : node
      )
    );

    setSelectedNode(null);
    setEditedNodeData(null);
    setParamsText('');
    setOptionsText('');
  };

  const handleCloseEditPanel = () => {
    setSelectedNode(null);
    setEditedNodeData(null);
    setParamsText('');
    setOptionsText('');
  };

  return (
    <div style={{ width: '100%', height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Editor Container */}
      <div style={{ display: 'flex', flex: 1, position: 'relative' }}>
        {/* Node Palette */}
        <NodePalette
          services={availableServices}
          onSave={handleSaveWorkflow}
          onClearCanvas={handleClearCanvas}
          onCancel={onCancel}
          onAutoLayout={handleAutoLayout}
          saving={saving}
          workflowName={workflowName}
          onWorkflowNameChange={setWorkflowName}
          isEditMode={!!workflowId}
        />

        {/* ReactFlow Canvas */}
        <div style={{ flex: 1 }}>
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onDragOver={onDragOver}
            onDrop={onDrop}
            onNodeClick={handleNodeClick}
            onEdgeClick={onEdgeClick}
            onReconnect={onReconnect}
            onReconnectStart={onReconnectStart}
            onReconnectEnd={onReconnectEnd}
            nodeTypes={nodeTypes}
            defaultEdgeOptions={{
              type: 'smoothstep',
              animated: true,
              style: { stroke: '#3498db', strokeWidth: 3 },
              markerEnd: { type: MarkerType.ArrowClosed, color: '#3498db' },
            }}
            defaultViewport={{ x: 0, y: 0, zoom: 0.6 }}
            minZoom={0.1}
            maxZoom={2}
            edgesReconnectable={true}
            fitView
            fitViewOptions={{
              padding: 0.2,
              maxZoom: 0.8,
            }}
          >
            <Background variant={BackgroundVariant.Dots} gap={20} size={1} />
            <Controls />
          </ReactFlow>
        </div>

        {/* Edit Panel - Right Sidebar */}
        {selectedNode && editedNodeData && (
          <div style={{
            width: '650px',
            background: 'white',
            borderLeft: '1px solid #e0e0e0',
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden',
          }}>
            {/* Panel Header */}
            <div style={{
              padding: '20px',
              borderBottom: '1px solid #e0e0e0',
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center',
            }}>
              <h3 style={{ margin: 0, fontSize: '18px', color: '#2c3e50' }}>
                Edit Node: {editedNodeData.label}
              </h3>
              <button
                onClick={handleCloseEditPanel}
                style={{
                  background: 'none',
                  border: 'none',
                  fontSize: '24px',
                  cursor: 'pointer',
                  color: '#999',
                }}
              >
                ×
              </button>
            </div>

            {/* Panel Body */}
            <div style={{
              flex: 1,
              overflowY: 'auto',
              padding: '20px',
            }}>
              <div style={{ marginBottom: '20px' }}>
                <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600, fontSize: '14px' }}>
                  Service Name
                </label>
                <input
                  type="text"
                  className="form-control"
                  value={editedNodeData.serviceName || ''}
                  onChange={(e) => setEditedNodeData({...editedNodeData, serviceName: e.target.value})}
                  disabled={editedNodeData.isParent}
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    fontSize: '14px',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                  }}
                />
              </div>

              <div style={{ marginBottom: '20px' }}>
                <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600, fontSize: '14px' }}>
                  Method
                </label>
                {(() => {
                  // Get available methods for this service
                  const serviceName = editedNodeData.serviceName;
                  let methods: string[] = [];

                  if (serviceName === 'QUEUE') {
                    methods = ['EnJob', 'evaluateJob'];
                  } else if (serviceName === 'MONGODB_WORKFLOW') {
                    methods = ['findById', 'find', 'findOne', 'create', 'update', 'delete'];
                  } else if (serviceName?.startsWith('MINIO')) {
                    methods = ['uploadBuffer', 'getFile', 'deleteFile', 'listFiles', 'generateTemplate1PDF'];
                  }

                  if (methods.length > 0) {
                    return (
                      <select
                        className="form-control"
                        value={editedNodeData.method || methods[0]}
                        onChange={(e) => setEditedNodeData({...editedNodeData, method: e.target.value})}
                        style={{
                          width: '100%',
                          padding: '8px 12px',
                          fontSize: '14px',
                          border: '1px solid #ddd',
                          borderRadius: '4px',
                        }}
                      >
                        {methods.map(method => (
                          <option key={method} value={method}>{method}</option>
                        ))}
                      </select>
                    );
                  }

                  // For dynamic services or unknown, use text input
                  return (
                    <input
                      type="text"
                      className="form-control"
                      value={editedNodeData.method || ''}
                      onChange={(e) => setEditedNodeData({...editedNodeData, method: e.target.value})}
                      style={{
                        width: '100%',
                        padding: '8px 12px',
                        fontSize: '14px',
                        border: '1px solid #ddd',
                        borderRadius: '4px',
                      }}
                    />
                  );
                })()}
              </div>

              {/* Code editor for evaluateJob */}
              {editedNodeData.method === 'evaluateJob' && (
                <div style={{ marginBottom: '20px' }}>
                  <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600, fontSize: '14px', color: '#e74c3c' }}>
                    Code (TypeScript/JavaScript) <span style={{ color: '#e74c3c' }}>*</span>
                  </label>
                  <div style={{
                    border: '2px solid #e74c3c',
                    borderRadius: '4px',
                    overflow: 'hidden',
                  }}>
                    <Editor
                      height="550px"
                      defaultLanguage="typescript"
                      value={editedNodeData.params?.code || ''}
                      onChange={(value) => {
                        const newParams = { ...(editedNodeData.params || {}) };
                        newParams.code = value || '';
                        setEditedNodeData({...editedNodeData, params: newParams});
                      }}
                      theme="vs-dark"
                      options={{
                        minimap: { enabled: false },
                        fontSize: 14,
                        lineNumbers: 'on',
                        roundedSelection: false,
                        scrollBeyondLastLine: false,
                        automaticLayout: true,
                        tabSize: 2,
                        wordWrap: 'on',
                        // Hover settings for better tooltip display
                        hover: {
                          enabled: true,
                          delay: 100,
                          sticky: true,
                        },
                        // Suggest settings
                        quickSuggestions: {
                          other: true,
                          comments: false,
                          strings: false,
                        },
                        suggestOnTriggerCharacters: true,
                        acceptSuggestionOnEnter: 'on',
                        // Better IntelliSense
                        parameterHints: {
                          enabled: true,
                        },
                      }}
                      beforeMount={(monaco) => {
                        // Configure TypeScript compiler options for better IntelliSense
                        monaco.languages.typescript.typescriptDefaults.setCompilerOptions({
                          target: monaco.languages.typescript.ScriptTarget.ES2020,
                          allowNonTsExtensions: true,
                          moduleResolution: monaco.languages.typescript.ModuleResolutionKind.NodeJs,
                          module: monaco.languages.typescript.ModuleKind.CommonJS,
                          noEmit: true,
                          esModuleInterop: true,
                          jsx: monaco.languages.typescript.JsxEmit.React,
                          allowJs: true,
                          typeRoots: ['node_modules/@types'],
                        });

                        // Configure diagnostics options
                        monaco.languages.typescript.typescriptDefaults.setDiagnosticsOptions({
                          noSemanticValidation: false,
                          noSyntaxValidation: false,
                        });

                        // Add custom type definitions for VM context
                        monaco.languages.typescript.typescriptDefaults.addExtraLib(
                          `
// VM Context Types - Available in evaluateJob sandbox
//
// IMPORTANT: Your code will be wrapped in an async function, so you can:
// - Use 'await' for async operations
// - Use 'return' to return values directly
// - Or set 'returnValue' variable
//
// Example:
//   const workflow = await findById(data.workflowId);
//   const result = await createJobWithChildren('QUEUE', 'EnJob', {}, []);
//   return result; // or: returnValue = result;

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

/**
 * Return value variable (optional)
 * You can either set this variable OR use 'return' statement
 */
declare let returnValue: any;

/**
 * Input data passed from job params
 */
declare const data: any;

/**
 * Console for logging
 */
declare const console: {
  log(...args: any[]): void;
  error(...args: any[]): void;
  warn(...args: any[]): void;
  info(...args: any[]): void;
};

/**
 * Date constructor
 */
declare const Date: DateConstructor;

/**
 * Generate a UUID v4 string
 */
declare function uuidv4(): string;

/**
 * Find a document by ID (async)
 * @param id - Document ID
 * @example
 * const workflow = await findById('123abc');
 */
declare function findById(id: any): Promise<any>;

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
declare function createJobWithChildren(
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
declare function addDataForWorkFlow(workFlow: IWorkflow, data: any): Promise<any>;
`,
                          'file:///vm-context.d.ts'
                        );
                      }}
                    />
                  </div>
                  <small style={{ color: '#e74c3c', fontSize: '11px', fontWeight: 600, display: 'block', marginTop: '4px' }}>
                    Required for evaluateJob method. TypeScript syntax checking enabled.
                  </small>
                </div>
              )}

              <div style={{ marginBottom: '20px' }}>
                <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600, fontSize: '14px' }}>
                  Parameters (JSON)
                </label>
                <textarea
                  className="form-control"
                  rows={10}
                  value={paramsText}
                  onChange={(e) => setParamsText(e.target.value)}
                  placeholder={editedNodeData.method === 'evaluateJob' ? '{"returnValue": "", "data": {}}' : '{"key": "value"}'}
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    fontSize: '13px',
                    fontFamily: 'monospace',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                  }}
                />
                {editedNodeData.method === 'evaluateJob' && (
                  <small style={{ color: '#666', fontSize: '11px', display: 'block', marginTop: '4px' }}>
                    Optional: returnValue, data, or other params (code is in separate field above)
                  </small>
                )}
              </div>

              <div style={{ marginBottom: '20px' }}>
                <label style={{ display: 'block', marginBottom: '8px', fontWeight: 600, fontSize: '14px' }}>
                  Options (BullMQ JobOptions - JSON)
                </label>
                <textarea
                  className="form-control"
                  rows={8}
                  value={optionsText}
                  onChange={(e) => setOptionsText(e.target.value)}
                  placeholder='{"delay": 5000, "priority": 1}'
                  style={{
                    width: '100%',
                    padding: '8px 12px',
                    fontSize: '13px',
                    fontFamily: 'monospace',
                    border: '1px solid #ddd',
                    borderRadius: '4px',
                  }}
                />
              </div>
            </div>

            {/* Panel Footer */}
            <div style={{
              padding: '20px',
              borderTop: '1px solid #e0e0e0',
              display: 'flex',
              gap: '10px',
              justifyContent: 'flex-end',
            }}>
              <button
                onClick={handleCloseEditPanel}
                className="btn btn-secondary"
              >
                Cancel
              </button>
              <button
                onClick={handleUpdateNodeData}
                className="btn btn-primary"
              >
                Save Changes
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

// Wrapper component with ReactFlowProvider
const WorkflowEditor: React.FC<WorkflowEditorProps> = (props) => {
  return (
    <ReactFlowProvider>
      <WorkflowEditorInner {...props} />
    </ReactFlowProvider>
  );
};

export default WorkflowEditor;
