CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    project_id UUID NOT NULL
        REFERENCES projects(id) ON DELETE CASCADE,

    assignee_id UUID,

    title VARCHAR(255) NOT NULL,
    description TEXT,

    status VARCHAR(20) NOT NULL DEFAULT 'TODO',
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',

    due_date DATE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_task_status
        CHECK (
            status IN (
                'TODO',
                'IN_PROGRESS',
                'REVIEW',
                'DONE'
            )
        ),

    CONSTRAINT chk_task_priority
        CHECK (
            priority IN ('LOW', 'MEDIUM', 'HIGH', 'URGENT')
        ),

    CONSTRAINT fk_tasks_assignee_membership
        FOREIGN KEY (project_id, assignee_id)
        REFERENCES project_members(project_id, user_id)
        ON DELETE RESTRICT
);

CREATE INDEX idx_tasks_project_status
    ON tasks(project_id, status);

CREATE INDEX idx_tasks_assignee_id
    ON tasks(assignee_id);

CREATE INDEX idx_tasks_due_date
    ON tasks(due_date);
