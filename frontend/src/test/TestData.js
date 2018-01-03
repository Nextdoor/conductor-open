import {Modes} from 'types/config';

const commit1 = {
  author_email: 'bob@nextdoor.com',
  author_name: 'Bob Bailey',
  id: '1',
  message: 'first commit',
  sha: 'abc',
  url: 'https://github.com/Nextdoor/conductor/commit/abc',
};

const commit2 = {
  author_email: 'joe@nextdoor.com',
  author_name: 'Joe Smith',
  id: '2',
  message: 'second commit',
  sha: 'def',
  url: 'https://github.com/Nextdoor/conductor/commit/def',
};

const commit3 = {
  author_email: 'joe@nextdoor.com',
  author_name: 'Joe Smith',
  id: '3',
  message: 'third commit',
  sha: 'ghi',
  url: 'https://github.com/Nextdoor/repo/commit/ghi',
};

const ticket1 = {
  id: '1',
  key: 'abc',
  assignee_email: commit1.author_email,
  assignee_name: commit1.author_name,
  url: 'https://atlassian.net/browse/REL-100',
  created_at: '2000-01-01T00:00:00Z',
  closed_at: null,
  deleted_at: null,
  commits: [commit1],
};

const ticket2 = {
  id: '2',
  key: 'def',
  assignee_email: commit2.author_email,
  assignee_name: commit2.author_name,
  url: 'https://atlassian.net/browse/REL-101',
  created_at: '2000-01-01T00:00:00Z',
  closed_at: null,
  deleted_at: null,
  commits: [commit2, commit2],
};

const phaseGroups = {
  id: '1',
  head_sha: commit3.sha,
  delivery: {
    id: '1',
    started_at: '2000-01-01T00:00:00Z',
    completed_at: null,
    type: 0,
    error: null,
    jobs: []
  },
  verification: {
    id: '2',
    started_at: null,
    completed_at: null,
    type: 1,
    error: null,
    jobs: []
  },
  deploy: {
    id: '3',
    started_at: null,
    completed_at: null,
    type: 2,
    error: null,
    jobs: []
  },
};

export const newTrain = {
  id: '2',
  previous_id: '1',
  next_id: null,

  created_at: '2000-01-01T00:00:00Z',
  deployed_at: null,

  branch: 'master',
  head_sha: commit3.sha,
  tail_sha: commit1.sha,

  closed: false,
  schedule_override: false,

  blocked: false,

  commits: [commit1, commit2, commit3],

  engineer: {
    id: '100',
    name: 'Rob Mackenzie',
    created_at: '2000-01-01T00:00:00Z',
    email: 'rob@nextdoor.com',
    avatar_url: null,
  },

  tickets: [ticket1, ticket2],

  active_phases: phaseGroups,

  all_phase_groups: [phaseGroups],

  active_phase: 0,

  last_delivered_sha: null,

  previous_train_done: false,

  not_deployable_reason: null,

  done: false,

  can_rollback: false
};

export const noRequest = {
  fetching: false,
  error: null,
  receivedAt: null
};

export const completeRequest = {
  fetching: false,
  receivedAt: 1492551647,
  error: null,
};

const configOptions = {
  close_time: [{
    every: [1, 2, 3, 4, 5],
    start_time: {hour: 0, minute: 0},
    end_time: {hour: 1, minute: 0}}
  ]};

export const configSchedule = {
  mode: Modes.Schedule,
  options: configOptions,
};

export const configManual = {
  mode: Modes.Manual,
  options: configOptions,
};

export const user = {
  is_admin: false,
  name: 'Regular User',
  email: 'user@domain.com',
  avatar_url: null,

};

export const adminUser = {
  is_admin: true,
  name: 'Admin User',
  email: 'admin@domain.com',
  avatar_url: null,
};
