import React from 'react';
import { Project } from '../types';
import { useNavigate } from 'react-router-dom';

interface Props {
  project: Project;
}

const ProjectCard: React.FC<Props> = ({ project }) => {
  const navigate = useNavigate();

  return (
    <div
      className="project-card"
      onClick={() => navigate(`/projects/${project.id}`)}
      id={`project-card-${project.id}`}
    >
      <div className="project-card-header">
        <div className="project-card-icon">📁</div>
        <h3 className="project-card-title">{project.name}</h3>
      </div>
      {project.description && (
        <p className="project-card-desc">{project.description}</p>
      )}
      <div className="project-card-footer">
        <span className="project-card-date">
          Created {new Date(project.created_at).toLocaleDateString()}
        </span>
      </div>
    </div>
  );
};

export default ProjectCard;
