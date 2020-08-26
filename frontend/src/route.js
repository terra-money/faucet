import React from 'react';
import { Route } from 'react-router-dom';
import Home from './Pages/Home';

export default function () {
  return (
    <React.Fragment>
      <Route exact path="/" component={Home} />
    </React.Fragment>
  );
}
