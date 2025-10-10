package scanner

import (
	"time"
)

type ScanProgress struct {
	Tasks []*ScanTask
}

type ScanTask struct {
	Id          any
	Description string
	Completed   bool
	StartTime   time.Time
	FinishTime  time.Time
}

func (t *ScanTask) Duration() time.Duration {
	if t.Completed {
		return t.FinishTime.Sub(t.StartTime)
	} else {
		return time.Since(t.StartTime)
	}
}

func (p *ScanProgress) AddTask(taskId any, desc string) {
	t := ScanTask{
		Id:          taskId,
		Description: desc,
		Completed:   false,
		StartTime:   time.Now(),
	}
	p.Tasks = append(p.Tasks, &t)
}

func (p *ScanProgress) CompleteTask(taskId any) {
	for _, t := range p.Tasks {
		if t.Id == taskId {
			t.Completed = true
			t.FinishTime = time.Now()
			break
		}
	}
}
