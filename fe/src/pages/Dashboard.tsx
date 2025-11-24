import React, { useEffect, useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { servicesApi } from '../api/services';
import { workflowsApi } from '../api/workflows';
import { queuesApi } from '../api/queues';
import { cronjobsApi } from '../api/cronjobs';

const Dashboard: React.FC = () => {
  const { user } = useAuth();
  const [stats, setStats] = useState({
    services: { total: 0, healthy: 0, enabled: 0 },
    workflows: { total: 0 },
    queues: { total: 0, active: 0, failed: 0 },
    cronjobs: { total: 0, enabled: 0 },
  });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const [servicesRes, workflowsRes, queuesRes, cronjobsRes] = await Promise.all([
          servicesApi.getAll(),
          workflowsApi.getAll(),
          queuesApi.getAll(),
          cronjobsApi.getAll(),
        ]);

        const services = servicesRes.data || [];
        const queues = queuesRes.data || [];

        setStats({
          services: {
            total: services.length,
            healthy: services.filter((s) => s.healthy).length,
            enabled: services.filter((s) => s.enabled).length,
          },
          workflows: {
            total: workflowsRes.data?.length || 0,
          },
          queues: {
            total: queues.length,
            active: queues.reduce((sum, q) => sum + q.jobCounts.active, 0),
            failed: queues.reduce((sum, q) => sum + q.jobCounts.failed, 0),
          },
          cronjobs: {
            total: cronjobsRes.data?.length || 0,
            enabled: cronjobsRes.data?.filter((c) => c.enabled).length || 0,
          },
        });
      } catch (error) {
        console.error('Failed to fetch stats:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchStats();
  }, []);

  if (loading) {
    return (
      <div className="loading-container">
        <div className="spinner"></div>
        <p>Loading dashboard...</p>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>Dashboard</h1>
        <p>Welcome back, {user?.email}!</p>
      </div>

      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-icon">⚙️</div>
          <div className="stat-content">
            <h3>Services</h3>
            <div className="stat-value">{stats.services.total}</div>
            <div className="stat-details">
              <span className="stat-label">Healthy: {stats.services.healthy}</span>
              <span className="stat-label">Enabled: {stats.services.enabled}</span>
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">🔄</div>
          <div className="stat-content">
            <h3>Workflows</h3>
            <div className="stat-value">{stats.workflows.total}</div>
            <div className="stat-details">
              <span className="stat-label">Total workflows</span>
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">📥</div>
          <div className="stat-content">
            <h3>Queues</h3>
            <div className="stat-value">{stats.queues.total}</div>
            <div className="stat-details">
              <span className="stat-label">Active Jobs: {stats.queues.active}</span>
              <span className="stat-label error">Failed: {stats.queues.failed}</span>
            </div>
          </div>
        </div>

        <div className="stat-card">
          <div className="stat-icon">⏰</div>
          <div className="stat-content">
            <h3>Cron Jobs</h3>
            <div className="stat-value">{stats.cronjobs.total}</div>
            <div className="stat-details">
              <span className="stat-label">Enabled: {stats.cronjobs.enabled}</span>
            </div>
          </div>
        </div>
      </div>

      <div className="dashboard-info">
        <div className="info-card">
          <h3>System Status</h3>
          <p className="status-ok">All systems operational</p>
          <p className="timestamp">Last updated: {new Date().toLocaleString()}</p>
        </div>

        <div className="info-card">
          <h3>Quick Actions</h3>
          <div className="quick-actions">
            <a href="/services" className="action-link">
              Manage Services
            </a>
            <a href="/workflows" className="action-link">
              Create Workflow
            </a>
            <a href="/queues" className="action-link">
              View Queues
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
