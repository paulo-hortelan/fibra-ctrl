import { BrowserRouter, Routes, Route } from "react-router-dom";
import Register from "./pages/Register.tsx";

function App() {
  return (
    <BrowserRouter>
      <div className="min-h-screen bg-gray-100">
        <Routes>
          <Route
            path="/"
            element={<div className="text-center p-4">Home Page</div>}
          />
          <Route path="/register" element={<Register />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}

export default App;