import cookie from 'react-cookie';

import Actions from 'types/actions';
import API from 'api';

const authCookieName = process.env.AUTH_COOKIE_NAME;

const set = (token) => {
  return {
    type: Actions.SetToken,
    token: token
  };
};

const promptLogin = () => {
  return {
    type: Actions.PromptLogin
  };
};

const get = () => (dispatch) => {
  const token = cookie.load(authCookieName);
  if (token === undefined) {
    return dispatch(promptLogin());
  } else {
    return dispatch(set(token));
  }
};

const del = () => {
  return {
    type: Actions.DeleteToken
  };
};

const requestLogout = () => {
  return {
    type: Actions.RequestLogout
  };
};

const receiveLogout = () => (dispatch) => {
  dispatch({
    type: Actions.ReceiveLogout,
    receivedAt: Date.now()
  });
  dispatch(promptLogin());
};

const receiveLogoutError = (error) => {
  return {
    type: Actions.ReceiveLogoutError,
    error: error,
    receivedAt: Date.now()
  };
};

const logout = () => (dispatch) => {
  API.logout(dispatch);
};

export default {
  set,
  promptLogin,
  get,
  del,
  requestLogout,
  receiveLogout,
  receiveLogoutError,
  logout
};
