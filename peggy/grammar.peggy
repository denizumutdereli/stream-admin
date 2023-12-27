start
  = policies:statement+ {
      return policies;
    }

statement
  = source:source " " action:action " " allowance:allowance "\n"? {
      return {
        source: source,
        action: action,
        allowance: allowance
      };
    }

source
  = "orders" / "order" / "users" / "kyc"

action
  = "read" / "update" / "delete" / "create"

allowance
  = "allowed" / "not allowed" / "require otp" / "ask permission" / "partial allowed"
