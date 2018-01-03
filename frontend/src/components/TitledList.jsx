import React from 'react';
import PropTypes from 'prop-types';

class TitledList extends React.Component {
  render() {
    const {className, items} = this.props;
    let fullClassName = 'titled-list';
    if (className) {
      fullClassName = 'titled-list ' + className;
    }
    return (
      <div className={fullClassName}>
        <div className="titled-list-titles">
          <ul>
            {items.map((item, i) =>
              <li key={i}>{item[0]}</li>
            )}
          </ul>
        </div>
        <div className="titled-list-items">
          <ul>
            {items.map((item, i) =>
              <li key={i}>{item[1]}</li>
            )}
          </ul>
        </div>
      </div>
    );
  }
}

TitledList.propTypes = {
  className: PropTypes.string,
  items: PropTypes.arrayOf(
    PropTypes.arrayOf(PropTypes.node)
  ).isRequired
};

export default TitledList;
