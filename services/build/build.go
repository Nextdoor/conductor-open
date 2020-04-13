/* Handles building jobs remotely (like Jenkins). */
package build

type Service interface {
	CancelJob(jobName string, jobURL string, params map[string]string) error
	TriggerJob(jobName string, params map[string]string) error
}
