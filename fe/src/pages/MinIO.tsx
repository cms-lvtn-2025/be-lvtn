import React, { useEffect, useState } from 'react';
import { minioApi } from '../api/minio';
import type { MinioConfig } from '../types';

const MinIO: React.FC = () => {
  const [configs, setConfigs] = useState<MinioConfig[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchConfigs();
  }, []);

  const fetchConfigs = async () => {
    try {
      const response = await minioApi.getAll();
      setConfigs(response.data || []);
    } catch (error) {
      console.error('Failed to fetch MinIO configs:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async (id: string) => {
    try {
      const response = await minioApi.testConnection(id);
      if (response.data?.connected) {
        const buckets = response.data.buckets.join(', ') || 'none';
        alert(`Connection successful! Buckets: ${buckets}`);
      } else {
        alert('Connection failed');
      }
      fetchConfigs();
    } catch (error: any) {
      alert(error.response?.data?.message || 'Connection test failed');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure?')) return;
    try {
      await minioApi.delete(id);
      fetchConfigs();
    } catch (error: any) {
      alert(error.response?.data?.message || 'Failed to delete config');
    }
  };

  if (loading) return <div className="loading-container"><div className="spinner"></div></div>;

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>MinIO Storage Management</h1>
      </div>
      <div className="table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Endpoint</th>
              <th>Port</th>
              <th>Bucket</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {configs.map((config) => (
              <tr key={config._id}>
                <td><strong>{config.name}</strong></td>
                <td>{config.endPoint}</td>
                <td>{config.port}</td>
                <td>{config.bucketName}</td>
                <td>
                  <span className={`health-badge ${config.connectionStatus?.connected ? 'healthy' : 'unhealthy'}`}>
                    {config.connectionStatus?.connected ? '✓ Connected' : '✗ Disconnected'}
                  </span>
                </td>
                <td className="actions">
                  <button onClick={() => handleTest(config._id)} className="btn-sm btn-primary">Test</button>
                  <button onClick={() => handleDelete(config._id)} className="btn-sm btn-danger">Delete</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default MinIO;
