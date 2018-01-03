/* Handles building jobs remotely (like Jenkins). */
package build

type Service interface {
	TriggerJob(jobName string, params map[string]string) error
}
