import React, { useEffect, useState } from 'react';
import { workflowsApi } from '../api/workflows';
import { cronjobsApi } from '../api/cronjobs';
import WorkflowEditor from '../components/WorkflowEditor';
import type { Workflow, CronJob } from '../types';

const Workflows: React.FC = () => {
  const [workflows, setWorkflows] = useState<Workflow[]>([]);
  const [loading, setLoading] = useState(true);
  const [showEditor, setShowEditor] = useState(false);
  const [editingWorkflowId, setEditingWorkflowId] = useState<string | undefined>();
  const [showCronModal, setShowCronModal] = useState(false);
  const [selectedWorkflow, setSelectedWorkflow] = useState<Workflow | null>(null);
  const [cronSchedule, setCronSchedule] = useState('*/5 * * * *');

  useEffect(() => {
    fetchWorkflows();
  }, []);

  const fetchWorkflows = async () => {
    try {
      const response = await workflowsApi.getAll();
      setWorkflows(response.data || []);
    } catch (error) {
      console.error('Failed to fetch workflows:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleExecute = async (id: string) => {
    try {
      await workflowsApi.execute(id);
      alert('Workflow executed successfully');
    } catch (error: any) {
      alert(error.response?.data?.message || 'Failed to execute workflow');
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this workflow?')) return;
    try {
      await workflowsApi.delete(id);
      fetchWorkflows();
      alert('Workflow deleted successfully');
    } catch (error: any) {
      alert(error.response?.data?.message || 'Failed to delete workflow');
    }
  };

  const handleCreateWorkflow = () => {
    setEditingWorkflowId(undefined);
    setShowEditor(true);
  };

  const handleEditWorkflow = (workflowId: string) => {
    setEditingWorkflowId(workflowId);
    setShowEditor(true);
  };

  const handleEditorClose = () => {
    setShowEditor(false);
    setEditingWorkflowId(undefined);
    fetchWorkflows();
  };

  // CronJob management functions
  const handleCreateCronJob = (workflow: Workflow) => {
    setSelectedWorkflow(workflow);
    setCronSchedule('*/5 * * * *');
    setShowCronModal(true);
  };

  const handleSaveCronJob = async () => {
    if (!selectedWorkflow) return;

    try {
      await cronjobsApi.create({
        cronExpression: cronSchedule,
        WL_id: selectedWorkflow._id,
        enabled: true,
      });

      alert('CronJob created successfully!');
      setShowCronModal(false);
      fetchWorkflows();
    } catch (error: any) {
      alert(error.response?.data?.message || 'Failed to create CronJob');
    }
  };

  const handleToggleCronJob = async (cronJob: CronJob) => {
    if (!confirm(`Are you sure you want to ${cronJob.enabled ? 'disable' : 'enable'} this CronJob?`)) return;

    try {
      await cronjobsApi.toggle(cronJob._id);
      alert(`CronJob ${cronJob.enabled ? 'disabled' : 'enabled'} successfully!`);
      fetchWorkflows();
    } catch (error: any) {
      alert(error.response?.data?.message || 'Failed to toggle CronJob');
    }
  };

  const handleDeleteCronJob = async (cronJobId: string) => {
    if (!confirm('Are you sure you want to delete this CronJob?')) return;

    try {
      await cronjobsApi.delete(cronJobId);
      alert('CronJob deleted successfully!');
      fetchWorkflows();
    } catch (error: any) {
      alert(error.response?.data?.message || 'Failed to delete CronJob');
    }
  };

  if (showEditor) {
    return (
      <div style={{ height: 'calc(100vh - 80px)' }}>
        <WorkflowEditor
          workflowId={editingWorkflowId}
          onSave={handleEditorClose}
          onCancel={handleEditorClose}
        />
      </div>
    );
  }

  if (loading) {
    return (
      <div className="loading-container">
        <div className="spinner"></div>
        <p>Loading workflows...</p>
      </div>
    );
  }

  return (
    <div>
      <div className="page-header">
        <h1>Workflows Management</h1>
        <p>Manage workflows and their associated cron jobs</p>
      </div>

      <div className="card">
        <div className="card-header">
          <h2>Actions</h2>
        </div>
        <div className="card-body">
          <div className="btn-group">
            <button onClick={handleCreateWorkflow} className="btn btn-primary">
              + Create New Workflow
            </button>
          </div>
        </div>
      </div>

      {workflows.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state-icon">📋</div>
          <h3>No Workflows Found</h3>
          <p>Create your first workflow to get started</p>
        </div>
      ) : (
        <div className="table-container" style={{ marginTop: '20px' }}>
          <table>
            <thead>
              <tr>
                <th>Workflow Name</th>
                <th>Parent Service</th>
                <th>Method</th>
                <th>Children</th>
                <th>CronJob Status</th>
                <th>Created At</th>
                <th>Workflow Actions</th>
                <th>CronJob Actions</th>
              </tr>
            </thead>
            <tbody>
              {workflows.map((workflow) => (
                <tr key={workflow._id}>
                  <td>
                    <strong>{workflow.name || 'Unnamed Workflow'}</strong>
                  </td>
                  <td>
                    <span className="badge badge-primary">{workflow.parentServiceName}</span>
                  </td>
                  <td>{workflow.parentMethod}</td>
                  <td>{workflow.children?.length || 0}</td>
                  <td>
                    {workflow.hasCronJob && workflow.cronJob ? (
                      <div style={{ textAlign: 'left' }}>
                        <span className={`badge ${workflow.cronJob.enabled ? 'badge-success' : 'badge-warning'}`}>
                          {workflow.cronJob.enabled ? '✓ Active' : '⏸ Paused'}
                        </span>
                        <br />
                        <small style={{ fontFamily: 'monospace', color: '#666', fontSize: '11px' }}>
                          {workflow.cronJob.schedule}
                        </small>
                      </div>
                    ) : (
                      <span className="badge badge-secondary">✗ No CronJob</span>
                    )}
                  </td>
                  <td>{new Date(workflow.createdAt).toLocaleString()}</td>
                  <td>
                    <div className="btn-group">
                      <button
                        onClick={() => handleEditWorkflow(workflow._id)}
                        className="btn btn-sm btn-primary"
                        style={{ background: '#f39c12' }}
                        title="Edit workflow"
                      >
                        ✏ Edit
                      </button>
                      <button
                        onClick={() => handleExecute(workflow._id)}
                        className="btn btn-sm btn-primary"
                        title="Execute workflow now"
                      >
                        ▶ Execute
                      </button>
                      <button
                        onClick={() => handleDelete(workflow._id)}
                        className="btn btn-sm btn-danger"
                        title="Delete workflow"
                      >
                        🗑 Delete
                      </button>
                    </div>
                  </td>
                  <td>
                    <div className="btn-group">
                      {workflow.hasCronJob && workflow.cronJob ? (
                        <>
                          <button
                            onClick={() => handleToggleCronJob(workflow.cronJob!)}
                            className={`btn btn-sm ${workflow.cronJob.enabled ? 'btn-icon-danger' : 'btn-icon-primary'}`}
                            title={workflow.cronJob.enabled ? 'Disable CronJob' : 'Enable CronJob'}
                          >
                            {workflow.cronJob.enabled ? '⏸ Disable' : '▶ Enable'}
                          </button>
                          <button
                            onClick={() => handleDeleteCronJob(workflow.cronJob!._id)}
                            className="btn btn-sm btn-danger"
                            title="Delete CronJob"
                          >
                            🗑
                          </button>
                        </>
                      ) : (
                        <button
                          onClick={() => handleCreateCronJob(workflow)}
                          className="btn btn-sm btn-primary"
                          style={{ background: '#27ae60' }}
                          title="Create CronJob"
                        >
                          + Create CronJob
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* CronJob Creation Modal */}
      {showCronModal && (
        <div className="modal-overlay" onClick={() => setShowCronModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Create CronJob for {selectedWorkflow?.name}</h3>
              <button className="modal-close" onClick={() => setShowCronModal(false)}>×</button>
            </div>
            <div className="modal-body">
              <div className="form-group">
                <label>Cron Schedule</label>
                <input
                  type="text"
                  className="form-control"
                  value={cronSchedule}
                  onChange={(e) => setCronSchedule(e.target.value)}
                  placeholder="*/5 * * * *"
                />
                <small style={{ color: '#666', fontSize: '12px', marginTop: '8px', display: 'block' }}>
                  Examples:
                  <br />
                  • <code>*/5 * * * *</code> - Every 5 minutes
                  <br />
                  • <code>0 * * * *</code> - Every hour
                  <br />
                  • <code>0 0 * * *</code> - Every day at midnight
                  <br />
                  • <code>0 0 * * 0</code> - Every Sunday at midnight
                </small>
              </div>
            </div>
            <div className="modal-footer">
              <button className="btn btn-secondary" onClick={() => setShowCronModal(false)}>
                Cancel
              </button>
              <button className="btn btn-primary" onClick={handleSaveCronJob}>
                Create CronJob
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Workflows;
