import React from 'react';
import {withRouter} from 'react-router';
import PropTypes from 'prop-types';

import {trainProps, requestProps} from 'types/proptypes';
import Error from 'components/Error';

class Header extends React.Component {
  componentWillMount() {
    const {request, load} = this.props;
    if (request.fetching !== true && request.receivedAt === null) {
      load();
    }
  }

  render() {
    return (
      <div>
        {this.getComponent()}
      </div>
    );
  }

  search(event) {
    const commit = event.target.value.trim();
    this.props.onSearchCommit && this.props.onSearchCommit(commit);
    if (commit.length > 0) {
      this.props.router.push('/search/commit/' + commit);
    } else {
      this.props.router.push('/');
    }
  }

  getComponent() {
    const {self, request, logout} = this.props;

    if (request.fetching !== true && request.receivedAt === null) {
      return null;
    }

    if (request.fetching === true || request.receivedAt === null) {
      return null;
    }

    let content = null;
    if (request.error !== null) {
      content = <Error message={request.error}/>;
    } else {
      content = (
        <span>
          <button className="header-logout-button" onClick={logout}>
            Logout
          </button>
          <div className="header-name header-name-long">
            Logged in as {self.name}
          </div>
          <div className="header-name header-name-short">
            {self.name}
          </div>
          <img className="header-avatar" src={self.avatar_url}/>
          <span className="header-search">
              <input type="text"
                          placeholder="Search trains by commit id"
                          autoFocus="true"
                          value={this.props.params.commit}
                          onChange={(event) => this.search(event)}/>
          </span>
        </span>
      );
    }

    return (
      <div className="header">
        <img className="header-brand-nextdoor-logo" src="/images/nextdoor.png"/>
        <div className="header-divider"/>
        <div className="header-brand">Conductor</div>
        {content}
      </div>
    );
  }
}

Header.propTypes = {
  self: PropTypes.shape({
    name: PropTypes.string.isRequired,
    email: PropTypes.string.isRequired,
    avatar_url: PropTypes.string.isRequired,
  }),
  request: requestProps.isRequired,
  train: trainProps,
  load: PropTypes.func.isRequired,
  logout: PropTypes.func.isRequired,
  onSearchCommit: PropTypes.func,
  router: PropTypes.element,
  params: PropTypes.element,
};

export default withRouter(Header);
