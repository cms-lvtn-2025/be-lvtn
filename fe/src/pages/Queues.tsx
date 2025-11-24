import React, { useMemo } from 'react';

const Queues: React.FC = () => {
  const bullmqUrl = useMemo(() => {
    const baseUrl = import.meta.env.VITE_API_URL?.replace('/api', '') || 'http://localhost:3000';
    const token = localStorage.getItem('accessToken');

    if (token) {
      return `${baseUrl}/bullmq?token=${encodeURIComponent(token)}`;
    }

    return `${baseUrl}/bullmq`;
  }, []);

  return (
    <div>
      <div className="page-header">
        <h1>BullMQ Queue Monitor</h1>
        <p>Manage and monitor all queue jobs</p>
      </div>

      <div className="card" style={{ marginTop: '20px' }}>
        <div className="card-body" style={{ padding: 0 }}>
          <iframe
            src={bullmqUrl}
            style={{
              width: '100%',
              height: 'calc(100vh - 200px)',
              border: 'none',
              minHeight: '600px'
            }}
            title="BullMQ Board"
          />
        </div>
      </div>
    </div>
  );
};

export default Queues;
