CREATE TABLE task_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    task_id UUID NOT NULL
        REFERENCES tasks(id) ON DELETE CASCADE,

    user_id UUID
        REFERENCES users(id) ON DELETE SET NULL,

    action VARCHAR(50) NOT NULL,
    metadata JSONB,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_task_activities_task_created
    ON task_activities(task_id, created_at DESC);
