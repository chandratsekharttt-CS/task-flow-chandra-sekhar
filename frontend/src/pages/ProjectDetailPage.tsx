import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import client from '../api/client';
import { ProjectWithTasks, Task, TaskStatus, CreateTaskPayload, UpdateTaskPayload } from '../types';
import { useAuth } from '../context/AuthContext';
import TaskCard from '../components/TaskCard';
import TaskModal from '../components/TaskModal';
import EmptyState from '../components/EmptyState';

const statusColumns: { key: TaskStatus; label: string; icon: string }[] = [
  { key: 'todo', label: 'Todo', icon: '📋' },
  { key: 'in_progress', label: 'In Progress', icon: '🔄' },
  { key: 'done', label: 'Done', icon: '✅' },
];

const ProjectDetailPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [project, setProject] = useState<ProjectWithTasks | null>(null);
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Filter state
  const [statusFilter, setStatusFilter] = useState<string>('all');

  // Modal state
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);

  // Edit project state
  const [editingProject, setEditingProject] = useState(false);
  const [editName, setEditName] = useState('');
  const [editDesc, setEditDesc] = useState('');

  const fetchProject = useCallback(async () => {
    try {
      setLoading(true);
      const res = await client.get<ProjectWithTasks>(`/projects/${id}`);
      setProject(res.data);
      setTasks(res.data.tasks || []);
      setError('');
    } catch (err: any) {
      if (err.response?.status === 404) {
        setError('Project not found');
      } else {
        setError(err.response?.data?.error || 'Failed to load project');
      }
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchProject();
  }, [fetchProject]);

  // ── Optimistic task status change ────────────────
  const handleStatusChange = async (taskId: string, newStatus: TaskStatus) => {
    const prev = [...tasks];
    // Optimistically update
    setTasks(tasks.map((t) => (t.id === taskId ? { ...t, status: newStatus } : t)));
    try {
      await client.patch(`/tasks/${taskId}`, { status: newStatus });
    } catch {
      // Revert on error
      setTasks(prev);
    }
  };

  // ── Create task ──────────────────────────────────
  const handleCreateTask = async (data: CreateTaskPayload | UpdateTaskPayload) => {
    await client.post(`/projects/${id}/tasks`, data);
    fetchProject();
  };

  // ── Update task ──────────────────────────────────
  const handleUpdateTask = async (data: CreateTaskPayload | UpdateTaskPayload) => {
    if (!editingTask) return;
    await client.patch(`/tasks/${editingTask.id}`, data);
    fetchProject();
  };

  // ── Delete task ──────────────────────────────────
  const handleDeleteTask = async (taskId: string) => {
    await client.delete(`/tasks/${taskId}`);
    fetchProject();
  };

  // ── Delete project ───────────────────────────────
  const handleDeleteProject = async () => {
    if (!window.confirm('Delete this project and all its tasks? This cannot be undone.')) return;
    try {
      await client.delete(`/projects/${id}`);
      navigate('/projects');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete project');
    }
  };

  // ── Edit project ─────────────────────────────────
  const startEditProject = () => {
    if (!project) return;
    setEditName(project.name);
    setEditDesc(project.description || '');
    setEditingProject(true);
  };

  const saveEditProject = async () => {
    try {
      await client.patch(`/projects/${id}`, {
        name: editName,
        description: editDesc || null,
      });
      setEditingProject(false);
      fetchProject();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to update project');
    }
  };

  // ── Filter tasks ─────────────────────────────────
  const filteredTasks = statusFilter === 'all'
    ? tasks
    : tasks.filter((t) => t.status === statusFilter);

  const tasksByStatus = (status: TaskStatus) =>
    filteredTasks.filter((t) => t.status === status);

  const isOwner = project?.owner_id === user?.id;

  if (loading) {
    return (
      <div className="page page-center">
        <div className="spinner" />
      </div>
    );
  }

  if (error && !project) {
    return (
      <div className="page">
        <div className="alert alert-error">{error}</div>
        <button className="btn btn-ghost" onClick={() => navigate('/projects')}>← Back to Projects</button>
      </div>
    );
  }

  if (!project) return null;

  return (
    <div className="page" id="project-detail-page">
      {/* Project Header */}
      <div className="project-detail-header">
        <button className="btn btn-ghost btn-sm" onClick={() => navigate('/projects')}>
          ← Back
        </button>

        {editingProject ? (
          <div className="project-edit-form">
            <input
              className="input"
              value={editName}
              onChange={(e) => setEditName(e.target.value)}
              placeholder="Project name"
              id="edit-project-name"
            />
            <textarea
              className="input textarea"
              value={editDesc}
              onChange={(e) => setEditDesc(e.target.value)}
              placeholder="Description"
              rows={2}
              id="edit-project-desc"
            />
            <div className="project-edit-actions">
              <button className="btn btn-primary btn-sm" onClick={saveEditProject}>Save</button>
              <button className="btn btn-ghost btn-sm" onClick={() => setEditingProject(false)}>Cancel</button>
            </div>
          </div>
        ) : (
          <div className="project-detail-info">
            <h1>{project.name}</h1>
            {project.description && <p className="project-detail-desc">{project.description}</p>}
            {isOwner && (
              <div className="project-detail-actions">
                <button className="btn btn-ghost btn-sm" onClick={startEditProject} id="edit-project-btn">✏️ Edit</button>
                <button className="btn btn-danger btn-sm" onClick={handleDeleteProject} id="delete-project-btn">🗑️ Delete</button>
              </div>
            )}
          </div>
        )}
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      {/* Toolbar */}
      <div className="project-toolbar">
        <div className="project-toolbar-left">
          <span className="project-task-count">{tasks.length} task{tasks.length !== 1 ? 's' : ''}</span>
          <select
            className="select-sm"
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value)}
            id="status-filter"
          >
            <option value="all">All Statuses</option>
            <option value="todo">Todo</option>
            <option value="in_progress">In Progress</option>
            <option value="done">Done</option>
          </select>
        </div>
        <button
          className="btn btn-primary btn-sm"
          onClick={() => { setEditingTask(null); setShowTaskModal(true); }}
          id="new-task-btn"
        >
          + Add Task
        </button>
      </div>

      {/* Kanban Board */}
      {tasks.length === 0 ? (
        <EmptyState
          icon="📝"
          title="No tasks yet"
          description="Create your first task to start tracking progress."
          action={
            <button className="btn btn-primary" onClick={() => { setEditingTask(null); setShowTaskModal(true); }}>
              + Create Task
            </button>
          }
        />
      ) : (
        <div className="kanban-board">
          {statusColumns.map((col) => {
            const colTasks = tasksByStatus(col.key);
            return (
              <div className={`kanban-column kanban-column-${col.key}`} key={col.key}>
                <div className="kanban-column-header">
                  <span className="kanban-column-icon">{col.icon}</span>
                  <span className="kanban-column-title">{col.label}</span>
                  <span className="kanban-column-count">{colTasks.length}</span>
                </div>
                <div className="kanban-column-body">
                  {colTasks.length === 0 && (
                    <div className="kanban-empty">No tasks</div>
                  )}
                  {colTasks.map((task) => (
                    <TaskCard
                      key={task.id}
                      task={task}
                      onEdit={(t) => { setEditingTask(t); setShowTaskModal(true); }}
                      onStatusChange={handleStatusChange}
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}

      <TaskModal
        isOpen={showTaskModal}
        onClose={() => { setShowTaskModal(false); setEditingTask(null); }}
        onSave={editingTask ? handleUpdateTask : handleCreateTask}
        onDelete={handleDeleteTask}
        task={editingTask}
        projectId={id || ''}
        isProjectOwner={isOwner}
        isAssignee={editingTask?.assignee_id === user?.id}
      />
    </div>
  );
};

export default ProjectDetailPage;
