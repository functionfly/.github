import { agentApi } from '@/api/agent';
import { getWalletInfo } from '@/api/billing';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ADMIN_DASHBOARD_URL, PLANS, ROUTES } from '@/lib/constants';
import { isPlatformAdminRole } from '@/lib/platform-admin';
import { cn } from '@/lib/utils';
import { useAuthStore } from '@/stores/authStore';
import { useQuery } from '@tanstack/react-query';
import {
  ChevronDown,
  CreditCard,
  DollarSign,
  LogOut,
  Settings,
  Shield,
  User,
  Wallet,
} from 'lucide-react';
import { useState } from 'react';
import { Link } from 'react-router-dom';

interface UserMenuProps {
  className?: string;
}

export function UserMenu({ className }: UserMenuProps) {
  const { user, logout } = useAuthStore();
  const [isOpen, setIsOpen] = useState(false);
  const walletAgentId =
    typeof window !== 'undefined' ? localStorage.getItem('ff-last-wallet-agent-id') : null;

  const { data: agentWalletData } = useQuery({
    // Include user id so switching accounts does not show the previous user's cached wallet.
    queryKey: ['agent-wallet-info', user?.id, walletAgentId],
    queryFn: async () => {
      if (!walletAgentId) return null;
      return agentApi.getWallet(walletAgentId);
    },
    enabled: !!user && !!walletAgentId,
    staleTime: 60 * 1000, // 1 minute
    retry: false,
  });

  const { data: billingWalletInfo } = useQuery({
    queryKey: ['billing-wallet-info', user?.id],
    queryFn: getWalletInfo,
    enabled: !!user,
    staleTime: 60 * 1000, // 1 minute
    retry: false,
  });

  if (!user) return null;

  const planInfo = PLANS[user.plan.toUpperCase() as keyof typeof PLANS] || PLANS.FREE;
  const isAdmin = isPlatformAdminRole(user.role);

  const handleLogout = () => {
    logout();
  };

  const getInitials = () => {
    // Prefer username, then name, then email as fallback
    const source = user.username || user.name || user.email;
    return (
      source
        .split(/[@.\s_-]+/)
        .filter(Boolean)
        .map((word) => word.charAt(0))
        .join('')
        .toUpperCase()
        .slice(0, 2) || '??'
    );
  };

  const formatBalance = (balance: number | undefined) => {
    if (balance === undefined || balance === null) return null;
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(balance);
  };

  const rawAgentWallet = agentWalletData?.wallet as unknown as Record<string, unknown> | undefined;
  const agentWalletInfo = rawAgentWallet
    ? {
        balance_usd: Number(rawAgentWallet.balance_usd ?? rawAgentWallet.balanceUSD ?? 0),
        total_earned_usd: Number(
          rawAgentWallet.total_earned_usd ?? rawAgentWallet.totalEarnedUSD ?? 0
        ),
        total_spent_usd: Number(
          rawAgentWallet.total_spent_usd ?? rawAgentWallet.totalSpentUSD ?? 0
        ),
      }
    : null;

  // Registry/billing top-ups credit the platform wallet; agent credits are separate. If localStorage
  // still points at an agent with $0, don't hide a non-zero billing wallet in the header.
  let walletInfo: typeof agentWalletInfo | typeof billingWalletInfo | null = null;
  let usingAgentWallet = false;
  if (agentWalletInfo && billingWalletInfo) {
    const agentBal = agentWalletInfo.balance_usd;
    const billBal = billingWalletInfo.balance_usd;
    if (billBal > 0 && agentBal <= 0) {
      walletInfo = billingWalletInfo;
      usingAgentWallet = false;
    } else if (agentBal > 0) {
      walletInfo = agentWalletInfo;
      usingAgentWallet = true;
    } else {
      walletInfo = billingWalletInfo;
      usingAgentWallet = false;
    }
  } else {
    walletInfo = agentWalletInfo ?? billingWalletInfo ?? null;
    usingAgentWallet = !!agentWalletInfo;
  }

  return (
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className={cn(
            'ml-2 flex items-center gap-3 pl-4 border-l border-border-subtle hover:bg-bg-hover',
            className
          )}
        >
          <div className="text-right hidden sm:block">
            <p className="text-sm font-medium text-text-primary truncate max-w-24">
              {user.username ? `@${user.username}` : user.name || user.email}
            </p>
            {walletInfo && (
              <p className="text-xs text-amber-500 flex items-center justify-end gap-1">
                <Wallet className="h-3 w-3" />
                {formatBalance(walletInfo.balance_usd)} balance
              </p>
            )}
            {!walletInfo && (
              <p className="text-xs text-text-muted capitalize">
                {user.name && user.username ? user.name : `${planInfo.name} Plan`}
              </p>
            )}
          </div>
          <div className="w-9 h-9 rounded-full bg-linear-to-br from-brand-500 to-brand-600 flex items-center justify-center text-white font-medium text-sm">
            {user.avatar ? (
              <img
                src={user.avatar}
                alt={user.name || user.username || 'User'}
                className="w-full h-full rounded-full object-cover"
              />
            ) : (
              getInitials()
            )}
          </div>
          <ChevronDown
            className={cn(
              'w-4 h-4 text-text-secondary transition-transform',
              isOpen && 'rotate-180'
            )}
          />
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="w-56 bg-bg-secondary border border-border-subtle shadow-xl"
        sideOffset={8}
      >
        <DropdownMenuLabel className="px-3 py-2">
          <div className="flex flex-col space-y-1">
            <p className="text-sm font-medium text-text-primary">
              {user.name || user.username || 'User'}
            </p>
            {user.username && <p className="text-xs text-brand-400">@{user.username}</p>}
            <p className="text-xs text-text-muted truncate">{user.email}</p>
          </div>
        </DropdownMenuLabel>

        {/* Wallet Balance Section */}
        {walletInfo && (
          <>
            <div className="px-3 py-2 mx-2 my-1 rounded-md bg-amber-500/10 border border-amber-500/20">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <DollarSign className="h-4 w-4 text-amber-500" />
                  <span className="text-xs text-text-secondary">
                    {usingAgentWallet ? 'Agent Wallet Balance' : 'Wallet Balance'}
                  </span>
                </div>
                <span className="text-sm font-semibold text-amber-500">
                  {formatBalance(walletInfo.balance_usd)}
                </span>
              </div>
              <div className="flex items-center justify-between mt-1 text-xs text-text-muted">
                <span>{usingAgentWallet ? 'Total Earned' : 'Earned'}</span>
                <span className="text-green-400">
                  {formatBalance(
                    usingAgentWallet
                      ? (walletInfo as { total_earned_usd?: number }).total_earned_usd
                      : (walletInfo as { lifetime_earnings_usd?: number }).lifetime_earnings_usd
                  )}
                </span>
              </div>
              <div className="flex items-center justify-between text-xs text-text-muted">
                <span>{usingAgentWallet ? 'Total Spent' : 'Fees Paid'}</span>
                <span className="text-red-400">
                  {formatBalance(
                    usingAgentWallet
                      ? (walletInfo as { total_spent_usd?: number }).total_spent_usd
                      : (walletInfo as { lifetime_fees_usd?: number }).lifetime_fees_usd
                  )}
                </span>
              </div>
            </div>
          </>
        )}

        <DropdownMenuSeparator className="bg-border-subtle" />

        <DropdownMenuItem asChild>
          <Link
            to={user.username ? `/u/${user.username}` : '/profile'}
            className="flex items-center gap-3 px-3 py-2 text-sm hover:bg-bg-hover cursor-pointer"
            onClick={() => setIsOpen(false)}
          >
            <User className="w-4 h-4" />
            <span>Profile</span>
          </Link>
        </DropdownMenuItem>

        <DropdownMenuItem asChild>
          <Link
            to={user.username ? `/u/${user.username}/settings` : ROUTES.SETTINGS}
            className="flex items-center gap-3 px-3 py-2 text-sm hover:bg-bg-hover cursor-pointer"
            onClick={() => setIsOpen(false)}
          >
            <Settings className="w-4 h-4" />
            <span>Settings</span>
          </Link>
        </DropdownMenuItem>

        <DropdownMenuItem asChild>
          <Link
            to={user.username ? `/u/${user.username}/settings/billing` : ROUTES.BILLING}
            className="flex items-center gap-3 px-3 py-2 text-sm hover:bg-bg-hover cursor-pointer"
            onClick={() => setIsOpen(false)}
          >
            <CreditCard className="w-4 h-4" />
            <span>Billing</span>
          </Link>
        </DropdownMenuItem>

        {isAdmin && ADMIN_DASHBOARD_URL && (
          <>
            <DropdownMenuSeparator className="bg-border-subtle" />
            <DropdownMenuItem asChild>
              <a
                href={ADMIN_DASHBOARD_URL}
                className="flex items-center gap-3 px-3 py-2 text-sm hover:bg-bg-hover cursor-pointer"
                onClick={() => setIsOpen(false)}
                rel="noopener noreferrer"
              >
                <Shield className="w-4 h-4" />
                <span>Admin Panel</span>
              </a>
            </DropdownMenuItem>
          </>
        )}

        <DropdownMenuSeparator className="bg-border-subtle" />

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
