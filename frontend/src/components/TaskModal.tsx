import React, { useState, useEffect } from 'react';
import { Task, CreateTaskPayload, UpdateTaskPayload, TaskStatus, TaskPriority } from '../types';
import client from '../api/client';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onSave: (data: CreateTaskPayload | UpdateTaskPayload) => Promise<void>;
  onDelete?: (taskId: string) => Promise<void>;
  task?: Task | null;          // null = create mode, Task = edit mode
  projectId: string;
  isProjectOwner?: boolean;
  isAssignee?: boolean;
}

const TaskModal: React.FC<Props> = ({ isOpen, onClose, onSave, onDelete, task, isProjectOwner = true, isAssignee = false }) => {
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [status, setStatus] = useState<TaskStatus>('todo');
  const [priority, setPriority] = useState<TaskPriority>('medium');
  const [dueDate, setDueDate] = useState('');
  const [assigneeId, setAssigneeId] = useState('');
  const [users, setUsers] = useState<any[]>([]);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

  const isEditMode = !!task;
  const isReadOnlyFields = isEditMode && !isProjectOwner && isAssignee;

  useEffect(() => {
    // Fetch users when modal opens
    if (isOpen) {
      client.get('/users').then((res: any) => setUsers(res.data)).catch(console.error);
    }
  }, [isOpen]);

  useEffect(() => {
    if (task) {
      setTitle(task.title);
      setDescription(task.description || '');
      setStatus(task.status);
      setPriority(task.priority);
      setDueDate(task.due_date ? task.due_date.split('T')[0] : '');
      setAssigneeId(task.assignee_id || '');
    } else {
      setTitle('');
      setDescription('');
      setStatus('todo');
      setPriority('medium');
      setDueDate('');
      setAssigneeId('');
    }
    setError('');
    setFieldErrors({});
  }, [task, isOpen]);

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setFieldErrors({});

    if (!title.trim()) {
      setFieldErrors({ title: 'Title is required' });
      return;
    }

    setSaving(true);
    try {
      const payload: CreateTaskPayload & UpdateTaskPayload = {
        title: title.trim(),
        description: description.trim() || undefined,
        status,
        priority,
        assignee_id: assigneeId || null,
        due_date: dueDate || null,
      };
      await onSave(payload);
      onClose();
    } catch (err: any) {
      if (err.response?.data?.fields) {
        setFieldErrors(err.response.data.fields);
      } else {
        setError(err.response?.data?.error || 'Something went wrong');
      }
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (!task || !onDelete) return;
    if (!window.confirm('Are you sure you want to delete this task?')) return;
    setSaving(true);
    try {
      await onDelete(task.id);
      onClose();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to delete task');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" onClick={(e) => e.stopPropagation()} id="task-modal">
        <div className="modal-header">
          <h2>{isEditMode ? 'Edit Task' : 'New Task'}</h2>
          <button className="btn-icon" onClick={onClose}>&times;</button>
        </div>

        <form onSubmit={handleSubmit} className="modal-body">
          {error && <div className="alert alert-error">{error}</div>}

          <div className="form-group">
            <label htmlFor="task-title">Title *</label>
            <input
              id="task-title"
              type="text"
              className={`input ${fieldErrors.title ? 'input-error' : ''}`}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              placeholder="Enter task title"
              disabled={isReadOnlyFields}
              autoFocus
            />
            {fieldErrors.title && <span className="field-error">{fieldErrors.title}</span>}
          </div>

          <div className="form-group">
            <label htmlFor="task-description">Description</label>
            <textarea
              id="task-description"
              className="input textarea"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Optional description"
              rows={3}
              disabled={isReadOnlyFields}
            />
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="task-status">Status</label>
              <select id="task-status" className="input" value={status} onChange={(e) => setStatus(e.target.value as TaskStatus)}>
                <option value="todo">Todo</option>
                <option value="in_progress">In Progress</option>
                <option value="done">Done</option>
              </select>
            </div>

            <div className="form-group">
              <label htmlFor="task-priority">Priority</label>
              <select id="task-priority" className="input" value={priority} onChange={(e) => setPriority(e.target.value as TaskPriority)} disabled={isReadOnlyFields}>
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </select>
            </div>
          </div>

          <div className="form-row">
            <div className="form-group">
              <label htmlFor="task-due-date">Due Date</label>
              <input
                id="task-due-date"
                type="date"
                className="input"
                value={dueDate}
                onChange={(e) => setDueDate(e.target.value)}
                disabled={isReadOnlyFields}
              />
            </div>
            
            <div className="form-group">
              <label htmlFor="task-assignee">Assignee</label>
              <select id="task-assignee" className="input" value={assigneeId} onChange={(e) => setAssigneeId(e.target.value)} disabled={isReadOnlyFields}>
                <option value="">Unassigned</option>
                {users.map(u => (
                  <option key={u.id} value={u.id}>{u.name}</option>
                ))}
              </select>
            </div>
          </div>

          <div className="modal-footer">
            {isEditMode && onDelete && isProjectOwner && (
              <button type="button" className="btn btn-danger" onClick={handleDelete} disabled={saving} id="task-delete-btn">
                Delete
              </button>
            )}
            <div className="modal-footer-right">
              <button type="button" className="btn btn-ghost" onClick={onClose} disabled={saving}>
                Cancel
              </button>
              <button type="submit" className="btn btn-primary" disabled={saving} id="task-save-btn">
                {saving ? 'Saving...' : isEditMode ? 'Save Changes' : 'Create Task'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};

export default TaskModal;
