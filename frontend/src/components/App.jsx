import React from 'react';
import PropTypes from 'prop-types';

import Auth from 'components/Auth';
import Header from 'containers/Header';
import Train from 'containers/Train';
import Search from 'containers/Search';

class App extends React.Component {
  componentWillMount() {
    const {needToken, promptLogin, getToken} = this.props;
    if (needToken === true && promptLogin !== true) {
      getToken();
    }
  }

  render() {
    const {needToken, promptLogin} = this.props;
    if (needToken === true && promptLogin !== true) {
      return null;
    }

    if (promptLogin) {
      return <Auth/>;
    }

    if (this.props.location.pathname.includes('/search')) {
      return this.getSearch(this.props.params);
    }

    return this.getTrain();
  }

  getTrain() {
    return (
      <div>
        <Header/>
        <Train trainId={this.props.params.trainId}/>
      </div>
    );
  }

  getSearch(params) {
    return (
      <div>
        <Header/>
        <Search params={params}/>
      </div>
    );
  }
}

App.propTypes = {
  needToken: PropTypes.bool.isRequired,
  promptLogin: PropTypes.bool.isRequired,
  getToken: PropTypes.func.isRequired,
  params: PropTypes.shape({
    trainId: PropTypes.string,

    search: PropTypes.bool,
    commit: PropTypes.string
  }),
  location: PropTypes.shape({
    pathname: PropTypes.string.isRequired
  }).isRequired
};

export default App;
