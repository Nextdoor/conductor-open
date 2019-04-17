import {combineReducers} from 'redux';

import admin from 'reducers/admin';
import phase from 'reducers/phase';
import search from 'reducers/search';
import self from 'reducers/self';
import token from 'reducers/token';
import train from 'reducers/train';

export const initialState = {
  admin: {
    config: null,
    requestConfig: {
      fetching: false,
      error: null,
      receivedAt: null
    },
    requestToggleMode: {
      fetching: false,
      error: null,
      receivedAt: null
    }
  },
  phase: {
    deployRequest: {
      fetching: false,
      error: null,
      receivedAt: null
    }
  },
  search: {
    details: null,
    request: {
      fetching: false,
      error: null,
      receivedAt: null
    }
  },
  self: {
    details: null,
    request: {
      fetching: false,
      error: null,
      receivedAt: null
    }
  },
  token: {
    promptLogin: false,
    token: null,
    logoutRequest: {
      fetching: false,
      error: null,
      receivedAt: null
    }
  },
  train: {
    details: null,
    request: {
      fetching: false,
      error: null,
      receivedAt: null
    },
    requestToggleClose: {
      fetching: false,
      error: null,
      receivedAt: null
    },
    requestExtend: {
      fetching: false,
      error: null,
      receivedAt: null
    },
    requestBlock: {
      fetching: false,
      error: null,
      receivedAt: null
    },
    requestUnblock: {
      fetching: false,
      error: null,
      receivedAt: null
    },
    requestCancel: {
      fetching: false,
      error: null,
      receivedAt: null
    },
    requestRollback: {
      fetching: false,
      error: null,
      receivedAt: null
    }
  }
};

export default combineReducers({
  admin,
  phase,
  search,
  self,
  token,
  train,
});
