# Page snapshot

```yaml
- generic [ref=e4]:
  - generic [ref=e5]:
    - heading "makoclaw" [level=1] [ref=e6]
    - paragraph [ref=e7]: AI Agent Control Panel
  - generic [ref=e8]:
    - heading "Login" [level=2] [ref=e9]
    - generic [ref=e10]:
      - generic [ref=e11]:
        - generic [ref=e12]: Email or Username
        - textbox "Email or Username" [ref=e13]:
          - /placeholder: Enter your email or username
          - text: admin
      - generic [ref=e14]:
        - generic [ref=e15]: Password
        - generic [ref=e16]:
          - textbox "Password" [ref=e17]:
            - /placeholder: Enter your password
            - text: kako9812claw
          - button "Show password" [ref=e18] [cursor=pointer]:
            - img [ref=e19]
      - generic [ref=e23]:
        - paragraph [ref=e24]: Login failed. Please try again.
        - button "Cerrar mensaje" [ref=e25] [cursor=pointer]
      - button "Sign In" [ref=e26] [cursor=pointer]
    - paragraph [ref=e27]:
      - text: Don't have an account?
      - link "Sign Up" [ref=e28] [cursor=pointer]:
        - /url: /signup
```