import React, { useContext } from 'react';

import { networks } from '../../../config';
import NetworkContext from '../../../contexts/NetworkContext';
import './NetworkSelector.scss';

const NetworkSelector = () => {
  const { network, setNetwork } = useContext(NetworkContext);

  const onChange = (e) => {
    setNetwork(e.target.value);
  };

  return (
    <div className="NetworkSelectorWrapper">
      <select className="NetworkSelector" value={network} onChange={onChange}>
        {networks.map(({ chainID, name }, index) => (
          <option value={chainID} key={index}>
            {name}
          </option>
        ))}
      </select>
    </div>
  );
};

export default NetworkSelector;
