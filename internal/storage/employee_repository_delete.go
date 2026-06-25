package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// DeleteEmployee deletes an employee by ID.
func (r *EmployeeRepository) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM employees WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("employee not found")
	}
	return nil
}

// DeleteDepartment deletes a department by ID.
func (r *EmployeeRepository) DeleteDepartment(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM departments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete department: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("department not found")
	}
	return nil
}

// RemoveEmployeeSkill removes a skill by ID.
func (r *EmployeeRepository) RemoveEmployeeSkill(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM employee_skills WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to remove employee skill: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("employee skill not found")
	}
	return nil
}

// DeleteProject deletes a project by ID.
func (r *EmployeeRepository) DeleteProject(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}
	return nil
}

// DeleteTask deletes a task by ID.
func (r *EmployeeRepository) DeleteTask(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// DeleteKnowledgeArticle deletes a knowledge article by ID.
func (r *EmployeeRepository) DeleteKnowledgeArticle(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM knowledge_articles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete knowledge article: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("knowledge article not found")
	}
	return nil
}
