import { useState } from "react";
import { Link } from "react-router-dom";
import {
  User,
  Settings,
  CreditCard,
  LogOut,
  ChevronDown,
  Shield
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/authStore";
import { cn } from "@/lib/utils";
import { PLANS, ROUTES } from "@/lib/constants";

interface UserMenuProps {
  className?: string;
}

export function UserMenu({ className }: UserMenuProps) {
  const { user, logout } = useAuthStore();
  const [isOpen, setIsOpen] = useState(false);

  if (!user) return null;

  const planInfo = PLANS[user.plan.toUpperCase() as keyof typeof PLANS] || PLANS.FREE;
  const isAdmin = user.role && ["super_admin", "support", "billing_admin", "developer_admin"].includes(user.role);

  const handleLogout = () => {
    logout();
  };

  const getInitials = () => {
    // Prefer username, then name, then email as fallback
    const source = user.username || user.name || user.email;
    return source
      .split(/[@.\s_-]+/)
      .filter(Boolean)
      .map(word => word.charAt(0))
      .join('')
      .toUpperCase()
      .slice(0, 2) || '??';
  };

  return (
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className={cn(
            "ml-2 flex items-center gap-3 pl-4 border-l border-white/8 hover:bg-white/5",
            className
          )}
        >
          <div className="text-right hidden sm:block">
            <p className="text-sm font-medium text-white truncate max-w-24">
              {user.username ? `@${user.username}` : (user.name || user.email)}
            </p>
            <p className="text-xs text-text-muted capitalize">
              {user.name && user.username ? user.name : `${planInfo.name} Plan`}
            </p>
          </div>
          <div className="w-9 h-9 rounded-full bg-linear-to-br from-brand-500 to-brand-600 flex items-center justify-center text-white font-medium text-sm">
            {user.avatar ? (
              <img src={user.avatar} alt={user.name || user.username || 'User'} className="w-full h-full rounded-full object-cover" />
            ) : (
              getInitials()
            )}
          </div>
          <ChevronDown className={cn(
            "w-4 h-4 text-text-secondary transition-transform",
            isOpen && "rotate-180"
          )} />
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="w-56 bg-bg-secondary border border-white/12 shadow-xl"
        sideOffset={8}
      >
        <DropdownMenuLabel className="px-3 py-2">
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium text-white">{user.name || user.username || 'User'}</p>
            {user.username && (
              <p className="text-xs text-brand-400">@{user.username}</p>
            )}
            <p className="text-xs text-text-muted truncate">{user.email}</p>
          </div>
        </DropdownMenuLabel>

        <DropdownMenuSeparator className="bg-white/8" />

        <DropdownMenuItem asChild>
          <Link
            to="/profile"
            className="flex items-center gap-3 px-3 py-2 text-sm hover:bg-white/5 cursor-pointer"
            onClick={() => setIsOpen(false)}
          >
            <User className="w-4 h-4" />
            <span>Profile</span>
          </Link>
        </DropdownMenuItem>

        <DropdownMenuItem asChild>
          <Link
            to={ROUTES.SETTINGS}
            className="flex items-center gap-3 px-3 py-2 text-sm hover:bg-white/5 cursor-pointer"
            onClick={() => setIsOpen(false)}
          >
            <Settings className="w-4 h-4" />
            <span>Settings</span>
          </Link>
        </DropdownMenuItem>

        <DropdownMenuItem asChild>
          <Link
            to="/billing"
            className="flex items-center gap-3 px-3 py-2 text-sm hover:bg-white/5 cursor-pointer"
            onClick={() => setIsOpen(false)}
          >
            <CreditCard className="w-4 h-4" />
            <span>Billing</span>
          </Link>
        </DropdownMenuItem>

        {isAdmin && (
          <>
            <DropdownMenuSeparator className="bg-white/8" />
            <DropdownMenuItem asChild>
              <Link
                to={ROUTES.ADMIN}
                className="flex items-center gap-3 px-3 py-2 text-sm hover:bg-white/5 cursor-pointer"
                onClick={() => setIsOpen(false)}
              >
                <Shield className="w-4 h-4" />
                <span>Admin Panel</span>
              </Link>
            </DropdownMenuItem>
          </>
        )}

        <DropdownMenuSeparator className="bg-white/8" />

        <DropdownMenuItem
          onClick={handleLogout}
          className="flex items-center gap-3 px-3 py-2 text-sm text-red-400 hover:bg-red-500/10 hover:text-red-300 cursor-pointer"
        >
          <LogOut className="w-4 h-4" />
          <span>Logout</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}