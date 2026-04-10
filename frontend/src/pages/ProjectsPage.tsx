import React, { useState, useEffect, useCallback } from 'react';
import client from '../api/client';
import { Project, PaginatedResponse } from '../types';
import ProjectCard from '../components/ProjectCard';
import CreateProjectModal from '../components/CreateProjectModal';
import EmptyState from '../components/EmptyState';

const ProjectsPage: React.FC = () => {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateModal, setShowCreateModal] = useState(false);

  const fetchProjects = useCallback(async () => {
    try {
      setLoading(true);
      const res = await client.get<PaginatedResponse<Project>>('/projects');
      setProjects(res.data.data || []);
      setError('');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load projects');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchProjects();
  }, [fetchProjects]);

  const handleCreateProject = async (name: string, description: string) => {
    await client.post('/projects', {
      name,
      description: description || null,
    });
    fetchProjects();
  };

  return (
    <div className="page" id="projects-page">
      <div className="page-header">
        <div>
          <h1>My Projects</h1>
          <p className="page-subtitle">Manage your projects and tasks</p>
        </div>
        <button
          className="btn btn-primary"
          onClick={() => setShowCreateModal(true)}
          id="new-project-btn"
        >
          + New Project
        </button>
      </div>

      {loading && (
        <div className="page-center">
          <div className="spinner" />
        </div>
      )}

      {error && !loading && (
        <div className="alert alert-error">{error}</div>
      )}

      {!loading && !error && projects.length === 0 && (
        <EmptyState
          icon="📁"
          title="No projects yet"
          description="Create your first project to get started with task management."
          action={
            <button className="btn btn-primary" onClick={() => setShowCreateModal(true)}>
              + Create Project
            </button>
          }
        />
      )}

      {!loading && projects.length > 0 && (
        <div className="projects-grid">
          {projects.map((project) => (
            <ProjectCard key={project.id} project={project} />
          ))}
        </div>
      )}

      <CreateProjectModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onCreate={handleCreateProject}
      />
    </div>
  );
};

export default ProjectsPage;
