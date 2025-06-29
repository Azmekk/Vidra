# Vidra Instructions.

## YTDLP Module:
This module is written in python. It is a rest api used to interface with yt-dlp functionality since most wrappers are finnicky or annoying, or lack functionality. It uses venv and is it's own separate module/docker container. More information on how to run it in `instructions.md` within the `Vidra.YTDLP` folder.

## Frontend Module

This module is written in React and uses the Mantine component library for building user interfaces. It is designed to be a single-page application (SPA) and communicates with the `Vidra.Backend` module via RESTful APIs. The frontend is responsible for rendering the user interface, handling user interactions, and displaying data fetched from the backend.

## Backend Module
This module is written in C# and uses the ASP.NET Core framework to build a RESTful API. It serves as the public backend for the Vidra application, handling requests from the frontend, processing data, and interacting with the YTDLP module.


# How to run the application:

1. **YTDLP Module**: 
   - Navigate to the `Vidra.YTDLP` directory.
   - Follow the instructions in `instructions.md` to set up and run the YTDLP module.

2. **Frontend Module**:
   - Navigate to the `Vidra.Frontend` directory.
   - Install the required dependencies using `pnpm install`.
   - Start the development server using `pnpm start`.

3. **Backend Module**:
   - Navigate to the `Vidra.Backend` directory.
   - Restore the NuGet packages using `dotnet restore`.
   - Start the ASP.NET Core application using `dotnet run`.

## Additional Notes:
In order to satisfy the cors requirements, it is suggested to use a browser extension such as "CORS Unblock" or setting up a reverse proxy so that the frontend and backend appear to be served from the same origin, which is how they are intended to be used.