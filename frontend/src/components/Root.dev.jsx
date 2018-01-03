import React, {Component} from 'react';
import {Provider} from 'react-redux';
import {createStore, applyMiddleware, compose} from 'redux';
import thunkMiddleware from 'redux-thunk';

import DevTools from 'components/DevTools';
import Routes from 'components/Routes';
import reducer, {initialState} from 'reducers';

const enhancer = compose(
  applyMiddleware(
    thunkMiddleware
  ),
  DevTools.instrument()
);

const store = createStore(
  reducer,
  initialState,
  enhancer
);

export default class Root extends Component {
  render() {
    return (
      <Provider store={store}>
        <div>
          <Routes/>
          <DevTools/>
        </div>
      </Provider>
    );
  }
}
