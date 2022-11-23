import React from 'react';

import './Header.scss';
import NetworkSelector from './NetworkSelector/NetworkSelector';

const Header = () => {
  return (
    <header className="AppHeader">
      <div className="FaucetLogo" />
      <NetworkSelector />
    </header>
  );
};

export default Header;
