import React, { useEffect, useState } from 'react';
import { servicesApi } from '../api/services';
import type { Service } from '../types';

const Services: React.FC = () => {
  const [services, setServices] = useState<Service[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editingService, setEditingService] = useState<Service | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    url: '',
    port: 50051,
    protocol: 'grpc' as 'grpc' | 'http' | 'https',
    protoPath: '',
    protoPackage: '',
    enabled: true,
  });

  useEffect(() => {
    fetchServices();
  }, []);

  const fetchServices = async () => {
    try {
      const response = await servicesApi.getAll();
      setServices(response.data || []);
    } catch (error) {
      console.error('Failed to fetch services:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingService) {
        await servicesApi.update(editingService._id, formData);
      } else {
        await servicesApi.create(formData);
      }
      setShowModal(false);
      resetForm();
      fetchServices();
    } catch (error: any) {
      alert(error.response?.data?.message || 'Failed to save service');
    }
  };

  const handleEdit = (service: Service) => {
    setEditingService(service);
    setFormData({
      name: service.name,
      url: service.url,
      port: service.port,
      protocol: service.protocol,
      protoPath: service.protoPath || '',
      protoPackage: service.protoPackage || '',
      enabled: service.enabled,
    });
    setShowModal(true);
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this service?')) return;
    try {
      await servicesApi.delete(id);
      fetchServices();
    } catch (error: any) {
      alert(error.response?.data?.message || 'Failed to delete service');
    }
  };

  const handleToggle = async (id: string) => {
    try {
      await servicesApi.toggle(id);
      fetchServices();
    } catch (error: any) {
      alert(error.response?.data?.message || 'Failed to toggle service');
    }
  };

  const handleHealthCheck = async (id: string) => {
    try {
      await servicesApi.healthCheck(id);
      fetchServices();
    } catch (error: any) {
      alert(error.response?.data?.message || 'Health check failed');
    }
  };

  const resetForm = () => {
    setEditingService(null);
    setFormData({
      name: '',
      url: '',
      port: 50051,
      protocol: 'grpc',
      protoPath: '',
      protoPackage: '',
      enabled: true,
    });
  };

  if (loading) {
    return <div className="loading-container"><div className="spinner"></div><p>Loading...</p></div>;
  }

  return (
    <div className="page-container">
      <div className="page-header">
        <h1>Services Management</h1>
        <button onClick={() => { resetForm(); setShowModal(true); }} className="btn btn-primary">
          + Add Service
        </button>
      </div>

      <div className="table-container">
        <table className="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>URL</th>
              <th>Port</th>
              <th>Protocol</th>
              <th>Status</th>
              <th>Health</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {services.map((service) => (
              <tr key={service._id}>
                <td><strong>{service.name}</strong></td>
                <td>{service.url}</td>
                <td>{service.port}</td>
                <td><span className="badge">{service.protocol}</span></td>
                <td>
                  <span className={`status-badge ${service.enabled ? 'enabled' : 'disabled'}`}>
                    {service.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </td>
                <td>
                  <span className={`health-badge ${service.healthy ? 'healthy' : 'unhealthy'}`}>
                    {service.healthy ? '✓ Healthy' : '✗ Unhealthy'}
                  </span>
                </td>
                <td className="actions">
                  <button onClick={() => handleEdit(service)} className="btn-icon" title="Edit">✏️</button>
                  <button onClick={() => handleToggle(service._id)} className="btn-icon" title="Toggle">
                    {service.enabled ? '⏸️' : '▶️'}
                  </button>
                  <button onClick={() => handleHealthCheck(service._id)} className="btn-icon" title="Health Check">🏥</button>
                  <button onClick={() => handleDelete(service._id)} className="btn-icon" title="Delete">🗑️</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div className="modal-overlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <h2>{editingService ? 'Edit Service' : 'Add Service'}</h2>
            <form onSubmit={handleSubmit}>
              <div className="form-group">
                <label>Name</label>
                <input value={formData.name} onChange={(e) => setFormData({ ...formData, name: e.target.value })} required />
              </div>
              <div className="form-group">
                <label>URL</label>
                <input value={formData.url} onChange={(e) => setFormData({ ...formData, url: e.target.value })} required />
              </div>
              <div className="form-group">
                <label>Port</label>
                <input type="number" value={formData.port} onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) })} required />
              </div>
              <div className="form-group">
                <label>Protocol</label>
                <select value={formData.protocol} onChange={(e) => setFormData({ ...formData, protocol: e.target.value as any })}>
                  <option value="grpc">gRPC</option>
                  <option value="http">HTTP</option>
                  <option value="https">HTTPS</option>
                </select>
              </div>
              <div className="form-group">
                <label>Proto Path</label>
                <input value={formData.protoPath} onChange={(e) => setFormData({ ...formData, protoPath: e.target.value })} />
              </div>
              <div className="form-group">
                <label>Proto Package</label>
                <input value={formData.protoPackage} onChange={(e) => setFormData({ ...formData, protoPackage: e.target.value })} />
              </div>
              <div className="form-group">
                <label>
                  <input type="checkbox" checked={formData.enabled} onChange={(e) => setFormData({ ...formData, enabled: e.target.checked })} />
                  Enabled
                </label>
              </div>
              <div className="modal-actions">
                <button type="button" onClick={() => setShowModal(false)} className="btn btn-outline">Cancel</button>
                <button type="submit" className="btn btn-primary">{editingService ? 'Update' : 'Create'}</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default Services;
