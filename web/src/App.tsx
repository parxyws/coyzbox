import { BrowserRouter } from "react-router-dom"
import AppRoutes from "@/app/routes.tsx"

export function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  )
}

export default App
