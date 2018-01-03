import React, {Component} from 'react';
import {Router, Route, browserHistory} from 'react-router';

import App from 'containers/App';

export default class Routes extends Component {
  render() {
    return (
      <Router history={browserHistory}>
        <Route path="/" component={App}>
          <Route path="/train/:trainId" component={App}/>
          <Route path="/search/commit/:commit" component={App}/>
        </Route>
      </Router>
    );
  }
}
