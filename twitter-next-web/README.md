# Twitter Clone Frontend

A modern, responsive Twitter/X clone built with Next.js 15, React 19, Tailwind CSS, and Shadcn/UI.

## Features

- **Authentication**: Secure sign-up and login flow.
- **Home Feed**: Real-time tweet feed with infinite scrolling.
- **Tweet Creation**: Create tweets with text and image support.
- **Interactions**: Like, Retweet, Reply to tweets.
- **Hashtags**: Hashtag highlighting and search.
- **Profile**: User profiles with edit functionality and follow system.
- **Explore**: Trending hashtags and user suggestions.
- **Notifications**: Real-time notification system.
- **Messages**: UI implementation for direct messages.
- **Responsive Design**: Mobile-first approach with bottom navigation for small screens.

## Tech Stack

- **Framework**: [Next.js 15](https://nextjs.org/) (App Router)
- **Language**: TypeScript
- **Styling**: [Tailwind CSS v4](https://tailwindcss.com/)
- **Components**: [Shadcn/UI](https://ui.shadcn.com/)
- **Icons**: [Lucide React](https://lucide.dev/)
- **State Management**: [Zustand](https://github.com/pmndrs/zustand)
- **Data Fetching**: [TanStack Query](https://tanstack.com/query/latest)
- **Forms**: [Zod](https://zod.dev/) validation

## Getting Started

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/yourusername/twitter-clone-azure.git
    cd twitter-clone-azure/twitter-frontend
    ```

2.  Install dependencies:
    ```bash
    npm install
    ```

3.  Set up environment variables:
    Create a `.env.local` file in the root of `twitter-frontend` and add:
    ```env
    NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
    ```

4.  Run the development server:
    ```bash
    npm run dev
    ```

5.  Open [http://localhost:3000](http://localhost:3000) in your browser.

## Project Structure

- `src/app`: Next.js App Router pages and layouts.
- `src/components`: Reusable UI components.
- `src/hooks`: Custom React hooks (data fetching, logic).
- `src/store`: Zustand stores for global state.
- `src/lib`: Utility functions and configuration.
- `src/types`: TypeScript type definitions.

## Deployment

This application is deployed via Azure Container Apps using GitHub Actions.
Please refer to the [Root README](../README.md#deployment-azure--github-actions) for detailed deployment instructions and CI/CD setup.
