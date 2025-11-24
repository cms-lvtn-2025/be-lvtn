import React from 'react';
import { Handle, Position } from 'reactflow';

interface WorkflowNodeProps {
  data: any;
  isConnectable?: boolean;
}

const WorkflowNode: React.FC<WorkflowNodeProps> = ({ data, isConnectable }) => {

  const getNodeColor = () => {
    if (data.isParent) return '#27ae60';
    if (data.serviceName === 'QUEUE') return '#3498db';
    if (data.serviceName === 'MONGODB_WORKFLOW') return '#9b59b6';
    if (data.serviceName?.startsWith('MINIO')) return '#e67e22';
    return '#34495e';
  };

  return (
    <>
      <div
        style={{
          background: 'white',
          border: `3px solid ${getNodeColor()}`,
          borderRadius: '8px',
          minWidth: '200px',
          boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
        }}
      >
        {/* Node Header */}
        <div
          style={{
            background: getNodeColor(),
            color: 'white',
            padding: '10px 12px',
            borderRadius: '5px 5px 0 0',
            fontWeight: 600,
            fontSize: '13px',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            gap: '8px',
          }}
        >
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <span>{data.isParent ? '🎯' : '📋'}</span>
            <span>{data.isParent ? 'Parent Job' : data.label}</span>
          </div>
          {!data.isParent && data.onDelete && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                data.onDelete();
              }}
              style={{
                background: 'transparent',
                border: 'none',
                color: 'white',
                fontSize: '20px',
                cursor: 'pointer',
                padding: '0',
                width: '20px',
                height: '20px',
                lineHeight: '20px',
                fontWeight: 'bold',
              }}
              title="Delete node"
            >
              ×
            </button>
          )}
        </div>

        {/* Node Body */}
        <div style={{ padding: '12px' }}>
          <div style={{ marginBottom: '6px' }}>
            <strong style={{ fontSize: '11px', color: '#666' }}>Service:</strong>
            <div style={{ fontSize: '13px', color: '#333' }}>{data.serviceName}</div>
          </div>
          <div style={{ marginBottom: '6px' }}>
            <strong style={{ fontSize: '11px', color: '#666' }}>Method:</strong>
            <div style={{ fontSize: '13px', color: '#333' }}>{data.method}</div>
          </div>
          <div style={{
            marginTop: '8px',
            fontSize: '11px',
            color: '#999',
            fontStyle: 'italic',
            textAlign: 'center'
          }}>
            Click to edit
          </div>
        </div>

        {/* Handles for connections */}
        <Handle
          type="target"
          position={Position.Left}
          style={{
            background: getNodeColor(),
            width: '16px',
            height: '16px',
            border: '3px solid white',
            boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
          }}
          isConnectable={isConnectable}
        />
        <Handle
          type="source"
          position={Position.Right}
          style={{
            background: getNodeColor(),
            width: '16px',
            height: '16px',
            border: '3px solid white',
            boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
          }}
          isConnectable={isConnectable}
        />
      </div>
    </>
  );
};

export default WorkflowNode;
