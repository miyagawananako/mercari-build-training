import { useState } from 'react';
import { ItemList } from '~/components/ItemList';
import { Listing } from '~/components/Listing';
import { styled } from 'styled-components';
import { themeColors } from '~/color';

function App() {
  // reload ItemList after Listing complete
  const [reload, setReload] = useState(true);
  return (
    <_Wrapper>
      <_TitleWrapper>
        <_p>Simple Mercari</_p>
      </_TitleWrapper>
      <div>
        <Listing onListingCompleted={() => setReload(true)} />
      </div>
      <div>
        <ItemList reload={reload} onLoadCompleted={() => setReload(false)} />
      </div>
    </_Wrapper>
  );
}

const _Wrapper = styled.div`
  text-align: center;
  background-color: ${themeColors.blue.b200};
`;

const _TitleWrapper = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: left;
  font-size: calc(10px + 1vmin);
  color: ${themeColors.monotone.m0};
`;

const _p = styled.p`
  font-weight: bold;
  font-size: 1.5rem;
`;

export default App;
