# accountable-holo 📊

Welcome to **Accountable Holo**! ✨ A sleek, modern desktop application designed to help small businesses manage their finances with ease. Keep track of income, expenses, and generate insightful reports to understand your financial health at a glance.

Built with Go and the Fyne toolkit for a beautiful, cross-platform user experience.

## ✨ Core Features

- **Transaction Management:** 💸 Record income and expenses effortlessly.
- **Category System:** 🏷️ Organize your transactions with customizable categories.
- **Account Tracking:** 🏦 Manage multiple financial accounts.
- **Financial Reporting:** 📈 Generate PDF and CSV reports to analyze your financial performance.
- **Cash Reconciliation:** 🧾 Easily reconcile your accounts to ensure your books are accurate.
- **Sleek UI:** 🎨 A clean, intuitive interface that's easy to navigate.

## 🚀 Getting Started

Follow these steps to get Accountable Holo running on your local machine.

### 1. Prerequisites

Make sure you have the following installed:

- [Go](https://golang.org/doc/install) (version 1.18 or newer)
- [Soda CLI](https://gobuffalo.io/en/docs/db/soda/installation/) for database migrations.

### 2. Clone the Repository

```bash
git clone https://github.com/nelsonmarro/accountable-holo.git
cd accountable-holo
```

### 3. Configuration

This application requires two configuration files. You'll need to create them from the provided examples.

#### a. Application Configuration

Create a `config.yml` file inside the `/config` directory by copying the example:

```bash
cp config/config.yaml.example config/config.yml
```

Now, open `config/config.yml` and edit the settings to match your environment. Pay special attention to the `storage.attachment_path`.

#### b. Database Configuration

Create a `database.yml` file in the project root by copying the example:

```bash
cp database.yml.example database.yml
```

Open `database.yml` and update the `user` and `password` fields with your PostgreSQL credentials.

### 4. Set Up the Database

With your `database.yml` configured, you can now create and migrate the database using the Soda CLI.

```bash
# Create the database
soda db create -e development

# Run all migrations
soda db migrate up -e development
```

## ▶️ Running the Application

Once the configuration and database setup are complete, you can run the application using a simple command from the `Makefile`.

```bash
make run-desktop-app
```

And that's it! 🎉 The Accountable Holo application window should appear on your screen.

## 🛠️ Available Commands

This project uses a `Makefile` to simplify common tasks. Here are the available commands:

| Command                  | Description                                                                |
| ------------------------ | -------------------------------------------------------------------------- |
| `make run-desktop-app`   | Runs the desktop application directly.                                     |
| `make build-desktop-app` | Builds the application executable into the `/build/desktop_app` directory. |
| `make help`              | Shows a list of all available commands.                                    |

---

Happy accounting! If you have any questions, feel free to open an issue.
