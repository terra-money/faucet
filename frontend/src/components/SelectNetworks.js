import React, { useContext } from 'react';

import s from './SelectNetworks.module.scss';
import { networks } from '../config';
import NetworkContext from '../contexts/NetworkContext';

const SelectNetworks = (props) => {
  const { network, setNetwork } = useContext(NetworkContext);

  return (
    <div className={props.className}>
      <select
        className={s.select}
        value={network}
        onChange={(e) => setNetwork(e.target.value)}
      >
        {networks.map(({ key }, index) => (
          <option key={index}>{key}</option>
        ))}
      </select>
      <div className={s.addon}>
        <i className="material-icons">arrow_drop_down</i>
      </div>
    </div>
  );
};

export default SelectNetworks;
