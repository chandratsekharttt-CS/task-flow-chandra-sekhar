import React from 'react';

interface Props {
  icon?: string;
  title: string;
  description: string;
  action?: React.ReactNode;
}

const EmptyState: React.FC<Props> = ({ icon = '📋', title, description, action }) => {
  return (
    <div className="empty-state">
      <span className="empty-state-icon">{icon}</span>
      <h3 className="empty-state-title">{title}</h3>
      <p className="empty-state-desc">{description}</p>
      {action && <div className="empty-state-action">{action}</div>}
    </div>
  );
};

export default EmptyState;
