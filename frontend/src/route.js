import React from 'react';
import { Route } from 'react-router-dom';
import Home from './Pages/Home';
import About from './Pages/About';
import Team from './Pages/Team/Team';

export default function() {
  return (
    <React.Fragment>
      <Route exact path="/" component={Home} />
      <Route exact path="/about" component={About} />
      <Route exact path="/team" component={Team} />
    </React.Fragment>
  );
}
