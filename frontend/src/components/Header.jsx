import React from 'react';
import {withRouter} from 'react-router';
import PropTypes from 'prop-types';

import {trainProps, requestProps} from 'types/proptypes';

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
    if (commit.length > 0) {
      this.props.router.push('/search/commit/' + commit);
    } else {
      this.props.router.push('/');
    }
  }

  getComponent() {
    const {self, request} = this.props;

    if (request.fetching !== true && request.receivedAt === null) {
      return null;
    }

    if (request.fetching === true || request.receivedAt === null) {
      return null;
    }

    let content = null;
    if (request.error !== null) {
      content = null;
    } else {
      content = (
        <span>
          <div className="header-name header-name-long">
            Logged in as {self.name}
          </div>
          <div className="header-name header-name-short">
            {self.name}
          </div>
          <div className="header-search">
              <input type="text"
                    placeholder="Search trains by commit sha"
                    autoFocus="true"
                    value={this.props.params.commit || ""}
                    onChange={(event) => this.search(event)}/>
          </div>
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
  }),
  request: requestProps.isRequired,
  train: trainProps,
  load: PropTypes.func.isRequired,
  router: PropTypes.object.isRequired,
  params: PropTypes.object.isRequired,
};

export default withRouter(Header);
