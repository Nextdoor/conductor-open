import PropTypes from 'prop-types';

export const commitProps = PropTypes.shape({
  author_email: PropTypes.string.isRequired,
  author_name: PropTypes.string.isRequired,
  id: PropTypes.string.isRequired,
  message: PropTypes.string.isRequired,
  sha: PropTypes.string.isRequired,
  url: PropTypes.string.isRequired
});

export const ticketProps = PropTypes.shape({
  id: PropTypes.string.isRequired,
  key: PropTypes.string.isRequired,
  assignee_email: PropTypes.string.isRequired,
  assignee_name: PropTypes.string.isRequired,
  url: PropTypes.string.isRequired,
  created_at: PropTypes.string.isRequired,
  closed_at: PropTypes.string,
  deleted_at: PropTypes.string,
  commits: PropTypes.arrayOf(commitProps)
});

export const jobProps = PropTypes.shape({
  id: PropTypes.string.isRequired,
  name: PropTypes.string.isRequired,
  started_at: PropTypes.string,
  completed_at: PropTypes.string,
  url: PropTypes.string,
  result: PropTypes.number,
  metadata: PropTypes.string,
});

export const phaseProps = PropTypes.shape({
  id: PropTypes.string.isRequired,
  started_at: PropTypes.string,
  completed_at: PropTypes.string,
  type: PropTypes.number.isRequired,
  error: PropTypes.string,
  jobs: PropTypes.arrayOf(jobProps)
});

export const phaseGroupProps = PropTypes.shape({
  id: PropTypes.string.isRequired,
  head_sha: PropTypes.string.isRequired,
  delivery: phaseProps,
  verification: phaseProps,
  deploy: phaseProps,
});

export const trainProps = PropTypes.shape({
  id: PropTypes.string.isRequired,
  previous_id: PropTypes.string,
  next_id: PropTypes.string,

  created_at: PropTypes.string.isRequired,
  deployed_at: PropTypes.string,
  cancelled_at: PropTypes.string,

  branch: PropTypes.string.isRequired,
  head_sha: PropTypes.string.isRequired,
  tail_sha: PropTypes.string.isRequired,

  closed: PropTypes.bool.isRequired,
  schedule_override: PropTypes.bool.isRequired,

  blocked: PropTypes.bool.isRequired,

  commits: PropTypes.arrayOf(commitProps).isRequired,

  engineer: PropTypes.shape({
    id: PropTypes.string.isRequired,
    name: PropTypes.string.isRequired,
    created_at: PropTypes.string.isRequired,
    email: PropTypes.string.isRequired,
    avatar_url: PropTypes.string,
  }),

  tickets: PropTypes.arrayOf(ticketProps),

  active_phases: phaseGroupProps.isRequired,
  all_phase_groups: PropTypes.arrayOf(phaseGroupProps).isRequired,

  active_phase: PropTypes.number.isRequired,

  last_delivered_sha: PropTypes.string,

  not_deployable_reason: PropTypes.string,

  done: PropTypes.bool.isRequired,

  previous_train_done: PropTypes.bool.isRequired,

  can_rollback: PropTypes.bool.isRequired,
});

export const configProps = PropTypes.shape({
  mode: PropTypes.number,
  options: PropTypes.shape({
    lock_time: PropTypes.arrayOf(PropTypes.shape({
      every: PropTypes.arrayOf(PropTypes.number),
      start_time: PropTypes.shape({
        hour: PropTypes.number,
        minute: PropTypes.number
      }),
      end_time: PropTypes.shape({
        hour: PropTypes.number,
        minute: PropTypes.number
      })
    })),
  }),
});

export const requestProps = PropTypes.shape({
  fetching: PropTypes.bool.isRequired,
  error: PropTypes.string,
  receivedAt: PropTypes.number,
  searchQuery: PropTypes.string
});

export const searchProps = PropTypes.shape({
  params: PropTypes.shape().isRequired,
  results: PropTypes.arrayOf(PropTypes.shape()).isRequired,
});
