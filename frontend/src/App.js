import React, { Component } from 'react';
import { BrowserRouter as Router, Route } from 'react-router-dom';
import './App.scss';
import SelectNetworks from './components/SelectNetworks';
import Home from './components/Home';

class App extends Component {
  showCurrentYear() {
    return new Date().getFullYear();
  }
  render() {
    return (
      <Router>
        <header>
          <SelectNetworks className="network_select" />
        </header>
        <Route exact path="/" component={Home} />
        <footer>
          &copy; 2019-{this.showCurrentYear()} <span>Terra</span>
        </footer>
      </Router>
    );
  }
}

export default App;
