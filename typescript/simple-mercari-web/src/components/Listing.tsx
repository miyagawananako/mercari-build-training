import { styled } from 'styled-components';
import { themeColors } from '~/color';
import { useListingForm } from '~/hooks/useListingForm';

interface Prop {
  onListingCompleted: () => void;
}

export const Listing = ({ onListingCompleted }: Prop) => {
  const { values, uploadImageRef, onValueChange, onFileChange, onSubmit } =
    useListingForm(onListingCompleted);

  return (
    <_Wrapper>
      <form onSubmit={onSubmit}>
        <_InputWrapper>
          <_InputText
            type="text"
            name="name"
            id="name"
            placeholder="name"
            onChange={onValueChange}
            required
            value={values.name}
          />
          <_InputText
            type="text"
            name="category"
            id="category"
            placeholder="category"
            onChange={onValueChange}
            value={values.category}
          />
          <_InputFile
            type="file"
            name="image"
            id="image"
            onChange={onFileChange}
            required
            ref={uploadImageRef}
          />
          <_SubmitButton type="submit">List this item</_SubmitButton>
        </_InputWrapper>
      </form>
    </_Wrapper>
  );
};

const _Wrapper = styled.div`
  min-height: 8vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: left;
  font-size: calc(10px + 1vmin);
  color: ${themeColors.monotone.m0};
`;

const _InputWrapper = styled.div`
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  padding: 10px;

  @media (max-width: 768px) {
    flex-direction: column;
    align-items: stretch;
  }
`;

const _InputText = styled.input`
  height: 30px;
  min-width: 200px;
  padding: 0 10px;
  border: 1px solid ${themeColors.monotone.m200};
  border-radius: 4px;

  @media (max-width: 768px) {
    width: 100%;
  }
`;

const _InputFile = styled.input`
  height: 30px;

  @media (max-width: 768px) {
    width: 100%;
  }

  &::file-selector-button {
    background-color: ${themeColors.blue.b400};
    color: ${themeColors.monotone.m0};
    border: none;
    padding: 5px 10px;
    border-radius: 5px;
    cursor: pointer;
  }

  &::file-selector-button:hover {
    background-color: ${themeColors.blue.b700};
  }
`;

const _SubmitButton = styled.button`
  height: 30px;
  padding: 0 20px;
  background-color: ${themeColors.blue.b400};
  color: ${themeColors.monotone.m0};
  border: none;
  border-radius: 4px;
  cursor: pointer;

  &:hover {
    background-color: ${themeColors.blue.b700};
  }

  @media (max-width: 768px) {
    width: 100%;
  }
`;
