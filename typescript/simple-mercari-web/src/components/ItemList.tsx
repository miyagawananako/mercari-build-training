import { useEffect, useState } from 'react';
import { Item, fetchItems } from '~/api';
import styled from 'styled-components';
import { themeColors } from '~/color';

interface Prop {
  reload: boolean;
  onLoadCompleted: () => void;
}

export const ItemList = ({ reload, onLoadCompleted }: Prop) => {
  const [items, setItems] = useState<Item[]>([]);
  useEffect(() => {
    const fetchData = () => {
      fetchItems()
        .then((data) => {
          console.debug('GET success:', data);
          setItems(data.items);
          onLoadCompleted();
        })
        .catch((error) => {
          console.error('GET error:', error);
        });
    };

    if (reload) {
      fetchData();
    }
  }, [reload, onLoadCompleted]);

  return (
    <_Wrapper>
      <_GridWrapper>
        {items?.map((item) => {
          return (
            <_ItemWrapper key={item.id}>
              <_Img src={import.meta.env.VITE_BACKEND_URL + '/' + item.image} />
              <p>
                <span>Name: {item.name}</span>
                <br />
                <span>Category: {item.category}</span>
              </p>
            </_ItemWrapper>
          );
        })}
      </_GridWrapper>
    </_Wrapper>
  );
};

const _Wrapper = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: left;
  background-color: ${themeColors.blue.b500};
`;

const _GridWrapper = styled.div`
  width: 80%;
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 10px;
  padding: 20px;
  padding-bottom: 300px;

  @media (max-width: 768px) {
    grid-template-columns: 1fr 1fr;
  }
`;

const _ItemWrapper = styled.div`
  min-height: 8vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: left;
  font-size: calc(10px + 1vmin);
  color: white;
  background-color: ${themeColors.blue.b400};
  padding: 10px;
`;

const _Img = styled.img`
  width: 100%;
  height: 100%;
  object-fit: cover;
  max-height: 100px;
  max-width: 100px;
`;
