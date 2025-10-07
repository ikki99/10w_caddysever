package api

import (
	"encoding/json"
	"net/http"

	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
)

// TasksHandler 获取任务列表
func TasksHandler(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query("SELECT id, name, command, schedule, is_loop, status, last_run FROM tasks ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		var lastRun *string
		if err := rows.Scan(&t.ID, &t.Name, &t.Command, &t.Schedule, &t.IsLoop, &t.Status, &lastRun); err != nil {
			continue
		}
		if lastRun != nil {
			t.LastRun = *lastRun
		}
		tasks = append(tasks, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// AddTaskHandler 添加任务
func AddTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t models.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	_, err := db.Exec("INSERT INTO tasks (name, command, schedule, is_loop, status) VALUES (?, ?, ?, ?, ?)",
		t.Name, t.Command, t.Schedule, t.IsLoop, "waiting")
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteTaskHandler 删除任务
func DeleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	
	db := database.GetDB()
	_, err := db.Exec("DELETE FROM tasks WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ExecuteTaskHandler 立即执行任务
func ExecuteTaskHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	
	db := database.GetDB()
	var command string
	err := db.QueryRow("SELECT command FROM tasks WHERE id=?", id).Scan(&command)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: 执行命令
	// exec.Command("cmd", "/C", command).Run()
	
	db.Exec("UPDATE tasks SET last_run=CURRENT_TIMESTAMP, status='success' WHERE id=?", id)

	w.WriteHeader(http.StatusOK)
}
