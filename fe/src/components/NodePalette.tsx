import React, { useState } from 'react';

interface NodePaletteProps {
  services: any;
  onClearCanvas?: () => void;
  onSave?: () => void;
  onCancel?: () => void;
  onAutoLayout?: () => void;
  saving?: boolean;
  workflowName?: string;
  onWorkflowNameChange?: (name: string) => void;
  isEditMode?: boolean;
}

const NodePalette: React.FC<NodePaletteProps> = ({ services, onClearCanvas, onSave, onCancel, onAutoLayout, saving, workflowName, onWorkflowNameChange, isEditMode }) => {
  const [expandedSections, setExpandedSections] = useState({
    fixed: false,
    static: false,
    dynamic: false,
  });

  const toggleSection = (section: 'fixed' | 'static' | 'dynamic') => {
    setExpandedSections(prev => ({
      ...prev,
      [section]: !prev[section]
    }));
  };

  const onDragStart = (event: React.DragEvent, nodeType: string, serviceName: string, method?: string) => {
    event.dataTransfer.setData('application/reactflow', nodeType);
    event.dataTransfer.setData('serviceName', serviceName);
    if (method) {
      event.dataTransfer.setData('method', method);
    }
    event.dataTransfer.effectAllowed = 'move';
  };

  if (!services) {
    return (
      <div style={{
        width: '250px',
        background: '#f8f9fa',
        borderRight: '1px solid #dee2e6',
        padding: '15px',
        overflowY: 'auto',
      }}>
        <p>Loading services...</p>
      </div>
    );
  }

  return (
    <div style={{
      width: '250px',
      background: '#f8f9fa',
      borderRight: '1px solid #dee2e6',
      padding: '15px',
      overflowY: 'auto',
      display: 'flex',
      flexDirection: 'column',
    }}>
      <h4 style={{ marginBottom: '10px', color: '#333', fontSize: '16px', fontWeight: 600 }}>
        {isEditMode ? 'Edit Workflow' : 'New Workflow'}
      </h4>

      {/* Workflow Name Input */}
      <div style={{ marginBottom: '15px' }}>
        <input
          type="text"
          value={workflowName}
          onChange={(e) => onWorkflowNameChange?.(e.target.value)}
          placeholder="Workflow name..."
          style={{
            width: '100%',
            padding: '8px 10px',
            fontSize: '13px',
            border: '2px solid #e0e0e0',
            borderRadius: '4px',
          }}
        />
      </div>

      {/* Action Buttons */}
      <div style={{ marginBottom: '15px', display: 'flex', flexDirection: 'column', gap: '6px' }}>
        <button
          onClick={onSave}
          disabled={saving}
          style={{
            padding: '8px 12px',
            background: '#3498db',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: saving ? 'not-allowed' : 'pointer',
            fontSize: '12px',
            fontWeight: 600,
            opacity: saving ? 0.6 : 1,
          }}
        >
          {saving ? 'Saving...' : '💾 Save'}
        </button>
        <button
          onClick={onAutoLayout}
          style={{
            padding: '8px 12px',
            background: '#9b59b6',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '12px',
            fontWeight: 600,
          }}
        >
          🎯 Auto Layout
        </button>
        <button
          onClick={onClearCanvas}
          style={{
            padding: '8px 12px',
            background: '#e74c3c',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '12px',
            fontWeight: 600,
          }}
        >
          🗑 Clear
        </button>
        <button
          onClick={onCancel}
          style={{
            padding: '8px 12px',
            background: '#95a5a6',
            color: 'white',
            border: 'none',
            borderRadius: '4px',
            cursor: 'pointer',
            fontSize: '12px',
            fontWeight: 600,
          }}
        >
          Cancel
        </button>
      </div>

      <p style={{ fontSize: '11px', color: '#666', marginBottom: '15px', padding: '8px', background: '#fff', borderRadius: '4px', border: '1px solid #e0e0e0' }}>
        💡 Drag nodes to canvas, click to edit
      </p>

      {/* Fixed Services */}
      <div style={{ marginBottom: '15px' }}>
        <h5
          onClick={() => toggleSection('fixed')}
          style={{
            fontSize: '13px',
            color: '#666',
            marginBottom: '10px',
            fontWeight: 600,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '8px',
            background: '#fff',
            borderRadius: '4px',
            border: '1px solid #e0e0e0',
          }}
        >
          <span>Fixed Services</span>
          <span style={{ fontSize: '16px' }}>{expandedSections.fixed ? '▼' : '▶'}</span>
        </h5>
        {expandedSections.fixed && services.fixed?.map((service: any) => (
          <div
            key={service.name}
            draggable
            onDragStart={(e) => onDragStart(e, 'workflow', service.name, service.methods[0])}
            style={{
              padding: '10px 12px',
              marginBottom: '8px',
              background: 'white',
              border: '2px solid #3498db',
              borderRadius: '6px',
              cursor: 'grab',
              fontSize: '13px',
              fontWeight: 500,
              display: 'flex',
              alignItems: 'center',
              gap: '8px',
              transition: 'all 0.2s',
            }}
            onMouseEnter={(e) => {
              e.currentTarget.style.transform = 'translateX(4px)';
              e.currentTarget.style.boxShadow = '0 2px 8px rgba(52, 152, 219, 0.2)';
            }}
            onMouseLeave={(e) => {
              e.currentTarget.style.transform = 'translateX(0)';
              e.currentTarget.style.boxShadow = 'none';
            }}
          >
            <span>📋</span>
            <span>{service.name}</span>
          </div>
        ))}
      </div>

      {/* Static Services (MinIO) */}
      {services.static && services.static.length > 0 && (
        <div style={{ marginBottom: '15px' }}>
          <h5
            onClick={() => toggleSection('static')}
            style={{
              fontSize: '13px',
              color: '#666',
              marginBottom: '10px',
              fontWeight: 600,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '8px',
              background: '#fff',
              borderRadius: '4px',
              border: '1px solid #e0e0e0',
            }}
          >
            <span>Static Services</span>
            <span style={{ fontSize: '16px' }}>{expandedSections.static ? '▼' : '▶'}</span>
          </h5>
          {expandedSections.static && services.static.map((service: any) => (
            <div
              key={service.name}
              draggable
              onDragStart={(e) => onDragStart(e, 'workflow', service.name)}
              style={{
                padding: '10px 12px',
                marginBottom: '8px',
                background: 'white',
                border: '2px solid #e67e22',
                borderRadius: '6px',
                cursor: 'grab',
                fontSize: '13px',
                fontWeight: 500,
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                transition: 'all 0.2s',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = 'translateX(4px)';
                e.currentTarget.style.boxShadow = '0 2px 8px rgba(230, 126, 34, 0.2)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'translateX(0)';
                e.currentTarget.style.boxShadow = 'none';
              }}
            >
              <span>💾</span>
              <span>{service.name.replace('MINIO_SERVICE_', 'MINIO_')}</span>
            </div>
          ))}
        </div>
      )}

      {/* Dynamic Services (gRPC) */}
      {services.dynamic && services.dynamic.length > 0 && (
        <div style={{ marginBottom: '15px' }}>
          <h5
            onClick={() => toggleSection('dynamic')}
            style={{
              fontSize: '13px',
              color: '#666',
              marginBottom: '10px',
              fontWeight: 600,
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
              padding: '8px',
              background: '#fff',
              borderRadius: '4px',
              border: '1px solid #e0e0e0',
            }}
          >
            <span>Dynamic Services (gRPC)</span>
            <span style={{ fontSize: '16px' }}>{expandedSections.dynamic ? '▼' : '▶'}</span>
          </h5>
          {expandedSections.dynamic && services.dynamic.map((service: any) => (
            <div
              key={service.name}
              draggable
              onDragStart={(e) => onDragStart(e, 'workflow', service.name)}
              style={{
                padding: '10px 12px',
                marginBottom: '8px',
                background: 'white',
                border: '2px solid #9b59b6',
                borderRadius: '6px',
                cursor: 'grab',
                fontSize: '13px',
                fontWeight: 500,
                display: 'flex',
                alignItems: 'center',
                gap: '8px',
                transition: 'all 0.2s',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.transform = 'translateX(4px)';
                e.currentTarget.style.boxShadow = '0 2px 8px rgba(155, 89, 182, 0.2)';
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.transform = 'translateX(0)';
                e.currentTarget.style.boxShadow = 'none';
              }}
            >
              <span>🔧</span>
              <span>{service.name}</span>
            </div>
          ))}
        </div>
      )}

      {/* Instructions */}
      <div style={{
        padding: '12px',
        background: '#fff3cd',
        border: '1px solid #ffc107',
        borderRadius: '6px',
        fontSize: '12px',
        color: '#856404',
        marginTop: '20px',
      }}>
        <strong>💡 Tip:</strong>
        <ul style={{ margin: '8px 0 0 0', paddingLeft: '20px' }}>
          <li>Drag nodes to canvas</li>
          <li>Connect nodes by dragging from handles</li>
          <li>Children connect to Parent from left to right</li>
          <li>Click "Edit" on nodes to configure params</li>
        </ul>
      </div>
    </div>
  );
};

export default NodePalette;
