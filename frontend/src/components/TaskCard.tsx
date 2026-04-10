import React from 'react';
import { Task, TaskStatus } from '../types';

interface Props {
  task: Task;
  onEdit: (task: Task) => void;
  onStatusChange: (taskId: string, newStatus: TaskStatus) => void;
  showProjectInfo?: boolean;
}

const priorityConfig: Record<string, { label: string; className: string }> = {
  high:   { label: 'High',   className: 'badge-danger' },
  medium: { label: 'Medium', className: 'badge-warning' },
  low:    { label: 'Low',    className: 'badge-info' },
};

const statusOptions: { value: TaskStatus; label: string }[] = [
  { value: 'todo', label: 'Todo' },
  { value: 'in_progress', label: 'In Progress' },
  { value: 'done', label: 'Done' },
];

const TaskCard: React.FC<Props> = ({ task, onEdit, onStatusChange, showProjectInfo }) => {
  const priority = priorityConfig[task.priority] || priorityConfig.medium;

  const formatDate = (dateStr: string | null) => {
    if (!dateStr) return null;
    const d = new Date(dateStr);
    return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
  };

  const isOverdue = task.due_date && task.status !== 'done' && new Date(task.due_date) < new Date();

  return (
    <div className="task-card" id={`task-card-${task.id}`}>
      <div className="task-card-header">
        <h4 className="task-card-title" onClick={() => onEdit(task)}>{task.title}</h4>
        <span className={`badge ${priority.className}`}>{priority.label}</span>
      </div>

      {task.description && (
        <p className="task-card-desc">{task.description}</p>
      )}

      {showProjectInfo && task.project_name && (
        <div className="task-card-project-info" style={{ fontSize: '0.85rem', color: 'var(--text-secondary)', marginBottom: '0.5rem', padding: '0.25rem 0' }}>
          <div><strong>Project:</strong> {task.project_name}</div>
          {task.project_owner_name && <div><strong>Assigned by:</strong> {task.project_owner_name}</div>}
        </div>
      )}

      {task.assignee_name && (
        <div className="task-card-assignee" style={{ fontSize: '0.85rem', color: 'var(--primary-color)', marginBottom: '0.5rem', fontWeight: 500 }}>
          👤 {task.assignee_name}
        </div>
      )}

      <div className="task-card-meta">
        {task.due_date && (
          <span className={`task-card-due ${isOverdue ? 'overdue' : ''}`}>
            📅 {formatDate(task.due_date)}
          </span>
        )}
      </div>

      <div className="task-card-actions">
        <select
          className="select-sm"
          value={task.status}
          onChange={(e) => onStatusChange(task.id, e.target.value as TaskStatus)}
          id={`task-status-select-${task.id}`}
        >
          {statusOptions.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        <button className="btn btn-ghost btn-xs" onClick={() => onEdit(task)} id={`task-edit-btn-${task.id}`}>
          ✏️ Edit
        </button>
      </div>
    </div>
  );
};

export default TaskCard;
