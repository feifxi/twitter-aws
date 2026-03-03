import { create } from 'zustand';

interface UIState {
    isSignInModalOpen: boolean;
    openSignInModal: () => void;
    closeSignInModal: () => void;

    isSignUpModalOpen: boolean;
    openSignUpModal: () => void;
    closeSignUpModal: () => void;
}

export const useUIStore = create<UIState>((set) => ({
    isSignInModalOpen: false,
    openSignInModal: () => set({ isSignInModalOpen: true }),
    closeSignInModal: () => set({ isSignInModalOpen: false }),

    isSignUpModalOpen: false,
    openSignUpModal: () => set({ isSignUpModalOpen: true }),
    closeSignUpModal: () => set({ isSignUpModalOpen: false }),
}));
