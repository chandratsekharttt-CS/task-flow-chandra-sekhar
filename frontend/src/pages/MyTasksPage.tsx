import React, { useState, useEffect, useCallback } from 'react';
import client from '../api/client';
import { Task, TaskStatus } from '../types';
import TaskCard from '../components/TaskCard';
import TaskModal from '../components/TaskModal';
import EmptyState from '../components/EmptyState';

const statusColumns: { key: TaskStatus; label: string; icon: string }[] = [
  { key: 'todo', label: 'Todo', icon: '📋' },
  { key: 'in_progress', label: 'In Progress', icon: '🔄' },
  { key: 'done', label: 'Done', icon: '✅' },
];

const MyTasksPage: React.FC = () => {
  const [tasks, setTasks] = useState<Task[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  // Modal state
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [editingTask, setEditingTask] = useState<Task | null>(null);

  const fetchTasks = useCallback(async () => {
    try {
      setLoading(true);
      const res = await client.get<Task[]>('/tasks/me');
      setTasks(res.data || []);
      setError('');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load tasks');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTasks();
  }, [fetchTasks]);

  // ── Optimistic task status change ────────────────
  const handleStatusChange = async (taskId: string, newStatus: TaskStatus) => {
    const prev = [...tasks];
    setTasks(tasks.map((t) => (t.id === taskId ? { ...t, status: newStatus } : t)));
    try {
      await client.patch(`/tasks/${taskId}`, { status: newStatus });
    } catch {
      setTasks(prev); // Revert on error
    }
  };

  // ── Update task ──────────────────────────────────
  const handleUpdateTask = async (data: any) => {
    if (!editingTask) return;
    await client.patch(`/tasks/${editingTask.id}`, data);
    fetchTasks();
  };

  const tasksByStatus = (status: TaskStatus) => tasks.filter((t) => t.status === status);

  if (loading) {
    return (
      <div className="page page-center">
        <div className="spinner" />
      </div>
    );
  }

  return (
    <div className="page">
      <div className="page-header">
        <h1>My Tasks</h1>
        <p>Manage all tasks assigned to you across projects.</p>
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      {/* Kanban Board */}
      {tasks.length === 0 ? (
        <EmptyState
          icon="📝"
          title="No tasks assigned"
          description="You have no tasks assigned to you right now."
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

      {/* Task Modal (Edit Only Mode for MyTasks Page) */}
      <TaskModal
        isOpen={showTaskModal}
        onClose={() => { setShowTaskModal(false); setEditingTask(null); }}
        onSave={handleUpdateTask}
        task={editingTask}
        projectId={editingTask ? editingTask.project_id : ''}
      />
    </div>
  );
};

export default MyTasksPage;
