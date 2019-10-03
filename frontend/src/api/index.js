import Actions from 'actions';
import {Modes, modeToString} from 'types/config';

const request = (fetchPromise) => {
  return fetchPromise
    .then((response) => {
      if (response.status !== 200) {
        return response;
      }
      return response.json();
    })
    .catch((ex) => {
      return ex;
    });
};

const get = (endpoint) => {
  const fetchPromise = fetch(endpoint, {
    credentials: 'same-origin'
  });
  return request(fetchPromise);
};

const post = (endpoint, payload) => {
  let postData = null;
  if (payload !== undefined) {
    postData = new URLSearchParams();
    Object.keys(payload).forEach((key) => {
      postData.set(key, payload[key]);
    });
  }
  const fetchPromise = fetch(endpoint, {
    credentials: 'same-origin',
    method: 'POST',
    body: postData,
  });
  return request(fetchPromise);
};

const encodeQueryParams = (data) => {
  const query = [];
  for (const d in data) {
    if (!data.hasOwnProperty(d)) {
      continue;
    }
    query.push(encodeURIComponent(d) + "=" + encodeURIComponent(data[d]));
  }
  return query.join("&");
};

const baseURI = '/api';

const train = function(trainID) {
  const uri = baseURI + '/train';
  if (trainID !== undefined) {
    return uri + '/' + trainID;
  }
  return uri;
};
const self = baseURI + '/user';
const logout = baseURI + '/auth/logout';
const closeTrain = function(trainID) {
  return baseURI + '/train/' + trainID + '/close';
};
const openTrain = function(trainID) {
  return baseURI + '/train/' + trainID + '/open';
};
const extendTrain = function(trainID) {
  return baseURI + '/train/' + trainID + '/extend';
};
const blockTrain = function(trainID) {
  return baseURI + '/train/' + trainID + '/block';
};
const unblockTrain = function(trainID) {
  return baseURI + '/train/' + trainID + '/unblock';
};
const cancelTrain = function(trainID) {
  return baseURI + '/train/' + trainID + '/cancel';
};
const rollbackToTrain = function(trainID) {
  return baseURI + '/train/' + trainID + '/rollback';
};
const search = function(params) {
  return baseURI + '/search?' + encodeQueryParams(params);
};
const restartJob = function(trainID, phaseName) {
  return baseURI + '/train/' + trainID + '/phase/' + phaseName + '/restart';
};
const changeEngineer = function(trainID) {
  return baseURI + '/train/' + trainID + '/changeEngineer';
};



const config = baseURI + '/config';
const mode = baseURI + '/mode';

const handleResponse = (response, dispatch, receive, receiveError) => {
  if (response instanceof SyntaxError) {
    dispatch(receiveError(response.message));
    return;
  }

  if (response.status !== undefined) {
    switch (response.status) {
      case 401:
        dispatch(Actions.Token.promptLogin());
        return;
      default:
        response.json().then((body) =>
          dispatch(receiveError(body.error))
        );
        return;
    }
  }
  dispatch(receive(response.result));
};

const API = {
  getSelf: (dispatch) => {
    dispatch(Actions.Self.request());
    get(self)
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Self.receive,
          Actions.Self.receiveError);
      });
  },

  getTrain: (trainId, dispatch, wipeCurrentTrain) => {
    dispatch(Actions.Train.request(wipeCurrentTrain));
    get(train(trainId))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Train.receive,
          Actions.Train.receiveError);
      });
  },

  getConfig: (dispatch) => {
    dispatch(Actions.Admin.requestConfig());
    get(config)
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Admin.receiveConfig,
          Actions.Admin.receiveConfigError);
      });
  },

  toggleMode: (trainId, currentMode, dispatch) => {
    dispatch(Actions.Admin.requestToggleMode());
    const newMode = currentMode === Modes.Schedule ? Modes.Manual : Modes.Schedule;
    post(mode, {"mode": modeToString(newMode)})
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Admin.receiveToggleMode,
          Actions.Admin.receiveToggleModeError);
      });
  },

  toggleClose: (trainId, currentlyClosed, dispatch) => {
    dispatch(Actions.Train.requestToggleClose());
    const endpoint = currentlyClosed ? openTrain : closeTrain;
    post(endpoint(trainId))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Train.receiveToggleClose,
          Actions.Train.receiveToggleCloseError);
      });
  },

  extendTrain: (trainId, dispatch) => {
    dispatch(Actions.Train.requestExtend());
    post(extendTrain(trainId))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Train.receiveExtend,
          Actions.Train.receiveExtendError);
      });
  },

  blockTrain: (trainId, dispatch) => {
    dispatch(Actions.Train.requestBlock());
    post(blockTrain(trainId))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Train.receiveBlock,
          Actions.Train.receiveBlockError);
      });
  },

  unblockTrain: (trainId, dispatch) => {
    dispatch(Actions.Train.requestUnblock());
    post(unblockTrain(trainId))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Train.receiveUnblock,
          Actions.Train.receiveUnblockError);
      });
  },
 
  cancelTrain: (trainId, dispatch) => {
    dispatch(Actions.Train.requestCancel());
    post(cancelTrain(trainId))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Train.receiveCancel,
          Actions.Train.receiveCancelError);
      });
  },

  rollbackToTrain: (trainId, dispatch) => {
    dispatch(Actions.Train.requestRollback());
    post(rollbackToTrain(trainId))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Train.receiveRollback,
          Actions.Train.receiveRollbackError);
      });
  },

  restartJob: (trainId, phaseName, dispatch) => {
    dispatch(Actions.Train.requestRestart());
    post(restartJob(trainId, phaseName))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Train.receiveRestart,
          Actions.Train.receiveRestartError);
      });
  },

  changeEngineer: (trainId, dispatch) => {
    dispatch(Actions.Train.requestChangeEngineer());
    post(changeEngineer(trainId))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Train.receiveChangeEngineer,
          Actions.Train.receiveChangeEngineerError);
      });
  },

  logout: (dispatch) => {
    dispatch(Actions.Token.requestLogout());
    post(logout)
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Token.receiveLogout,
          Actions.Token.receiveLogoutError);
      });
  },

  getSearch: (dispatch, params) => {
    dispatch(Actions.Search.request());
    get(search(params))
      .then((response) => {
        handleResponse(
          response,
          dispatch,
          Actions.Search.receive,
          Actions.Search.receiveError);
      });
  },

  encodeQueryParams
};

export default API;
