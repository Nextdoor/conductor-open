import React from 'react';
import PropTypes from 'prop-types';

import Header from 'containers/Header';
import Train from 'containers/Train';
import Search from 'containers/Search';

class App extends React.Component {
  constructor(props) {
    super(props);
  }

  render() {
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
