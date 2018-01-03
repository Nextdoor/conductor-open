import API from 'api';
import Actions from 'types/actions';
import {stringToMode} from 'types/config';

const requestConfig = () => {
  return {
    type: Actions.RequestConfig
  };
};

const receiveConfig = (response) => {
  return {
    type: Actions.ReceiveConfig,
    config: response,
    receivedAt: Date.now()
  };
};

const receiveConfigError = (error) => {
  return {
    type: Actions.ReceiveConfigError,
    error: error,
    receivedAt: Date.now()
  };
};

const requestToggleMode = () => {
  return {
    type: Actions.RequestToggleMode
  };
};

const receiveToggleMode = (mode) => {
  return {
    type: Actions.ReceiveToggleMode,
    mode: stringToMode(mode),
    receivedAt: Date.now()
  };
};

const receiveToggleModeError = (error) => {
  return {
    type: Actions.ReceiveToggleModeError,
    error: error,
    receivedAt: Date.now()
  };
};

const fetchConfig = () => (dispatch) => {
  API.getConfig(dispatch);
};

const toggleMode = () => (dispatch, getState) => {
  const state = getState();
  const train = state.train.details;
  API.toggleMode(train.id, state.admin.config.mode, dispatch);
};

export default {
  requestConfig,
  receiveConfig,
  receiveConfigError,
  requestToggleMode,
  receiveToggleMode,
  receiveToggleModeError,
  fetchConfig,
  toggleMode
};
