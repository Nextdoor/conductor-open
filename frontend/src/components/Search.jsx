import React from 'react';
import PropTypes from 'prop-types';
import _ from 'lodash';

import {searchProps, requestProps} from 'types/proptypes';

import Card from 'components/Card';
import Error from 'components/Error';
import Loading from 'components/Loading';

class Search extends React.Component {
  constructor(props) {
    super(props);

    // Debounce search to avoid throttling the api
    this.searchDebounced = _.debounce(this.props.search, 300);
  }

  componentWillMount() {
    const {request, search, params} = this.props;
    if (request.fetching !== true && request.receivedAt === null) {
      search(params);
    }
  }

  componentWillReceiveProps(nextProps) {
    if (nextProps.params.commit !== this.props.params.commit) {
      this.searchDebounced(nextProps.params);
    }
  }

  render() {
    const {request, details, params} = this.props;

    // Request is still being fetched; render nothing
    if (request.fetching === true) {
      return null;
    }

    // Request might be fetched, but for some reason receivedAt is null; render nothing
    if (request.receivedAt === null) {
      return null;
    }

    // Don't render for previous commit
    if (request.searchQuery && (request.searchQuery !== params.commit)) {
      return null;
    }

    if (request.error !== null) {
      return <Error message={request.error}/>;
    }

    if (details === null) {
      return <Loading/>;
    }

    return (
      <Card header="Search Results">
        {this.getComponent()}
      </Card>
    );
  }

  getComponent() {
    const {details, params} = this.props;
    const trains = [];
    details.results.forEach(function(train) {
      trains.push(<TrainLink key={train.id} id={train.id}/>);
    });

    return (
      <div className="search-results">
        <div className="search-results-header">
          Results for <span className="sha-flat">{params.commit}</span>
        </div>
        {trains}
      </div>
    );
  }
}

Search.propTypes = {
  details: searchProps,
  params: PropTypes.shape().isRequired,
  request: requestProps.isRequired,
  search: PropTypes.func.isRequired,
  commit: PropTypes.string,
};

class TrainLink extends React.Component {
  render() {
    const {id} = this.props;

    return (
      <a href={"/train/" + id}>
        <button className="train-link">
          Train {id}
        </button>
      </a>
    );
  }
}

TrainLink.propTypes = {
  id: PropTypes.string.isRequired,
};

export default Search;
