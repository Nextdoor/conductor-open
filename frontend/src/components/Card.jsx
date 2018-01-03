import React from 'react';
import PropTypes from 'prop-types';

class Card extends React.Component {
  render() {
    const {header, className} = this.props;
    let fullClassName = 'card';
    if (className) {
      fullClassName = 'card ' + className;
    }
    return (
      <div className={fullClassName}>
        <div className="card-header">
          {header}
        </div>
        <div className="card-divider"/>
        <div className="card-body">
          {this.props.children}
        </div>
      </div>
    );
  }
}

Card.propTypes = {
  header: PropTypes.node.isRequired,
  className: PropTypes.string,
  children: PropTypes.node.isRequired
};

export default Card;
