import React, {Component} from 'react';
import {Provider} from 'react-redux';
import {createStore, applyMiddleware} from 'redux';
import thunkMiddleware from 'redux-thunk';

import Routes from 'components/Routes';
import reducer, {initialState} from 'reducers';

const store = createStore(
  reducer,
  initialState,
  applyMiddleware(
    thunkMiddleware
  )
);

export default class Root extends Component {
  render() {
    return (
      <Provider store={store}>
        <Routes/>
      </Provider>
    );
  }
}
