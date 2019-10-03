/* eslint-disable no-invalid-this */

import React from 'react';
import {mount} from 'enzyme';
import Commits from 'components/Commits';
import {newTrain, completeRequest} from 'test/TestData';

describe('Commits', function() {

  it('gets groups', function() {

    this.train = newTrain;
    this.wrapper = mount(
      <Commits train={this.train} request={completeRequest}/>);

    const groups = this.wrapper.instance().getGroups();
    expect(groups).toEqual(expect.arrayContaining([
      [this.train.commits[0].author_name,
        [{
          message: this.train.commits[0].message,
          url: this.train.commits[0].url,
        }],
        [{
          done: false,
          key: this.train.tickets[0].key,
          url: this.train.tickets[0].url,
        }]
      ],
      [this.train.commits[1].author_name,
        [{
          message: this.train.commits[1].message,
          url: this.train.commits[1].url,
        }, {
          message: this.train.commits[2].message,
          url: this.train.commits[2].url,
        }],
        [{
          done: false,
          key: this.train.tickets[1].key,
          url: this.train.tickets[1].url,
        }]
      ]
    ]));
    expect(groups.length).toEqual(2);
  });

  it('renders correctly', function() {

    this.train = newTrain;
    this.wrapper = mount(
      <Commits train={this.train} request={completeRequest}/>);

    expect(this.wrapper.text()).toEqual(expect.stringContaining(this.train.commits[0].author_name));
    expect(this.wrapper.text()).toEqual(expect.stringContaining(this.train.commits[0].message));
  });
});
