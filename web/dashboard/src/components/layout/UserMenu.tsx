import { agentApi } from '@/api/agent';
import { createCheckoutSession, getWalletInfo } from '@/api/billing';
import { teamsApi } from '@/api/teams';
import { PlanSelectionModal } from '@/components/enterprise';
import { LanguagePicker } from '@/components/common/LanguagePicker';
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
import { useTheme } from '@/components/common/ThemeProvider';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { AnimatePresence, motion } from 'framer-motion';
import * as React from 'react';
import {
  Building2,
  Check,
  ChevronDown,
  ChevronRight,
  CreditCard,
  Globe,
  LogOut,
  Plus,
  Settings,
  Shield,
  ShieldCheck,
  User,
  Wallet,
  Crown,
  TrendingUp,
  ArrowUpRight,
} from 'lucide-react';
import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

interface UserMenuProps {
  className?: string;
}

export function UserMenu({ className }: UserMenuProps) {
  const { t } = useTranslation();
  const { user, logout, mfaRequired } = useAuthStore();
  const { theme } = useTheme();
  const navigate = useNavigate();
  const [isOpen, setIsOpen] = useState(false);
  const [showTeamsSubmenu, setShowTeamsSubmenu] = useState(false);
  const [showPlanModal, setShowPlanModal] = useState(false);
  const [isCheckoutLoading, setIsCheckoutLoading] = useState(false);
  const walletAgentId =
    typeof window !== 'undefined' ? localStorage.getItem('ff-last-wallet-agent-id') : null;

  const { data: teamsData } = useQuery({
    queryKey: ['user-teams'],
    queryFn: () => teamsApi.list(),
    enabled: !!user && isOpen,
    staleTime: 5 * 60 * 1000,
  });

  const teams = teamsData?.teams || [];
  const currentTeam = teams[0];

  useEffect(() => {
    if (!isOpen) {
      setShowTeamsSubmenu(false);
    }
  }, [isOpen]);

  const { data: agentWalletData } = useQuery({
    queryKey: ['agent-wallet-info', user?.id, walletAgentId],
    queryFn: async () => {
      if (!walletAgentId) return null;
      return agentApi.getWallet(walletAgentId);
    },
    enabled: !!user && !!walletAgentId,
    staleTime: 60 * 1000,
    retry: false,
  });

  const { data: billingWalletInfo } = useQuery({
    queryKey: ['billing-wallet-info', user?.id],
    queryFn: getWalletInfo,
    enabled: !!user,
    staleTime: 60 * 1000,
    retry: false,
  });

  if (!user) return null;

  const planInfo = PLANS[user.plan.toUpperCase() as keyof typeof PLANS] || PLANS.FREE;
  const isAdmin = isPlatformAdminRole(user.role);
  const isDark = theme === 'dark';

  // Velocity brand colors
  const brand = {
    primary: '#FF6B35',
    gradient: 'linear-gradient(135deg, #FF6B35 0%, #FF8C42 50%, #FF4F5E 100%)',
    light: '#FF8C42',
    dark: '#E55A2B',
    glow: isDark ? 'rgba(255, 107, 53, 0.3)' : 'rgba(255, 107, 53, 0.2)',
  };

  const handleLogout = () => {
    logout();
  };

  const getInitials = () => {
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

  // Theme-aware colors
  const colors = {
    menuBg: isDark ? '#151520' : '#ffffff',
    cardBg: isDark ? '#1e1e2e' : '#f8f9fa',
    hoverBg: isDark ? '#252538' : '#e8eaed',
    border: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.08)',
    textPrimary: isDark ? '#f8f8fc' : '#1a1a1f',
    textSecondary: isDark ? '#a0a0b8' : '#5f6368',
    textMuted: isDark ? '#6e6e8a' : '#9aa0a6',
  };

  return (
    <>
    <DropdownMenu open={isOpen} onOpenChange={setIsOpen}>
      <DropdownMenuTrigger asChild>
        <Button
          variant="ghost"
          className={cn(
            'ml-2 flex items-center gap-3 pl-4 pr-2 py-1.5 h-auto border-l border-border-subtle/60 hover:bg-bg-hover/80 relative group transition-all duration-200',
            isOpen && 'bg-bg-hover/60',
            className
          )}
        >
          <div className="text-right hidden sm:block">
            <p className="text-sm font-semibold text-text-primary truncate max-w-28 leading-tight">
              {user.name || user.username || 'User'}
            </p>
            {walletInfo ? (
              <p className="text-xs text-emerald-500 flex items-center justify-end gap-1 mt-0.5 font-medium">
                <Wallet className="h-3 w-3" />
                {formatBalance(walletInfo.balance_usd)}
              </p>
            ) : (
              <p className="text-xs text-text-muted capitalize mt-0.5">
                {planInfo.name} Plan
              </p>
            )}
          </div>
          <div className="relative">
            <div 
              className={cn(
                "w-9 h-9 rounded-full flex items-center justify-center text-white font-semibold text-sm relative overflow-hidden transition-all duration-200",
                "ring-2 ring-offset-2 ring-offset-bg-primary ring-transparent",
                "group-hover:ring-[#FF6B35]/40",
                isOpen && "ring-[#FF6B35]/60"
              )}
              style={{ background: brand.gradient }}
            >
              {user.avatar ? (
                <img
                  src={user.avatar}
                  alt={user.name || user.username || 'User'}
                  className="w-full h-full rounded-full object-cover"
                />
              ) : (
                <span className="relative z-10">{getInitials()}</span>
              )}
              <div className="absolute inset-0 bg-gradient-to-tr from-black/20 to-transparent" />
            </div>
            <span className={cn(
              "absolute -bottom-0.5 -right-0.5 w-3.5 h-3.5 rounded-full border-2 border-bg-primary",
              user?.isOnline ? 'bg-emerald-500' : 'bg-slate-500'
            )}>
              {user?.isOnline && (
                <span className="absolute inset-0 rounded-full bg-emerald-500 animate-ping opacity-40" />
              )}
            </span>
            {!mfaRequired && (
              <span className="absolute -top-0.5 -right-0.5 w-4 h-4 rounded-full bg-emerald-500 border-2 border-bg-primary flex items-center justify-center" title="2FA Enabled">
                <ShieldCheck className="w-2 h-2 text-white" />
              </span>
            )}
          </div>
          <ChevronDown
            className={cn(
              'w-4 h-4 text-text-secondary transition-transform duration-200',
              isOpen && 'rotate-180 text-[#FF6B35]'
            )}
          />
        </Button>
      </DropdownMenuTrigger>

      <DropdownMenuContent
        align="end"
        className="w-72 p-0 overflow-hidden"
        sideOffset={12}
        style={{
          backgroundColor: colors.menuBg,
          border: `1px solid ${colors.border}`,
          boxShadow: isDark 
            ? `0 25px 50px -12px rgba(0, 0, 0, 0.7), 0 0 0 1px ${brand.primary}25`
            : `0 25px 50px -12px rgba(0, 0, 0, 0.15), 0 0 0 1px ${brand.primary}15`,
        }}
      >
        {/* Header */}
        <div className="relative overflow-hidden">
          <div className="absolute inset-0" style={{ background: `linear-gradient(to bottom right, ${brand.primary}10, transparent)` }} />
          <DropdownMenuLabel className="relative px-4 py-4 m-0">
            <div className="flex items-start gap-3.5">
              <div className="relative shrink-0">
                <div 
                  className="w-12 h-12 rounded-full flex items-center justify-center text-white font-semibold text-base relative overflow-hidden shadow-lg"
                  style={{ 
                    background: brand.gradient,
                    boxShadow: `0 10px 15px -3px ${brand.glow}`
                  }}
                >
                  {user.avatar ? (
                    <img
                      src={user.avatar}
                      alt={user.name || user.username || 'User'}
                      className="w-full h-full rounded-full object-cover"
                    />
                  ) : (
                    <span className="relative z-10">{getInitials()}</span>
                  )}
                  <div className="absolute inset-0 bg-gradient-to-tr from-black/20 to-transparent" />
                </div>
                <span className={cn(
                  "absolute -bottom-1 -right-1 w-3.5 h-3.5 rounded-full border-2",
                  user?.isOnline ? 'bg-emerald-500' : 'bg-slate-500'
                )}
                style={{ borderColor: colors.menuBg }}>
                  {user?.isOnline && (
                    <span className="absolute inset-0 rounded-full bg-emerald-500 animate-ping opacity-40" />
                  )}
                </span>
              </div>
              
              <div className="flex flex-col min-w-0 flex-1">
                <p className="text-sm font-bold truncate leading-tight" style={{ color: colors.textPrimary }}>
                  {user.name || user.username || 'User'}
                </p>
                {user.username && (
                  <p className="text-xs truncate font-medium" style={{ color: brand.primary }}>@{user.username}</p>
                )}
                <p className="text-xs truncate mt-0.5" style={{ color: colors.textMuted }}>{user.email}</p>
                
                <div className="flex items-center gap-2 mt-2">
                  <span className={cn(
                    "inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-medium border",
                    user?.isOnline 
                      ? isDark ? 'bg-emerald-500/15 text-emerald-400 border-emerald-500/25' : 'bg-emerald-100 text-emerald-700 border-emerald-200'
                      : isDark ? 'bg-slate-500/15 text-slate-400 border-slate-500/25' : 'bg-slate-100 text-slate-600 border-slate-200'
                  )}>
                    <span className={cn("w-1 h-1 rounded-full", user?.isOnline && 'bg-emerald-500 animate-pulse')} />
                    {user?.isOnline ? 'Online' : 'Away'}
                  </span>
                  {!mfaRequired && (
                    <span 
                      className={cn(
                        "inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium border",
                        isDark ? 'border-[#FF6B35]/25' : 'border-[#FF6B35]/20'
                      )}
                      style={{ 
                        backgroundColor: isDark ? `${brand.primary}15` : `${brand.primary}10`,
                        color: brand.primary 
                      }}
                    >
                      <ShieldCheck className="w-3 h-3" />
                      2FA
                    </span>
                  )}
                </div>
              </div>
            </div>
          </DropdownMenuLabel>
        </div>

        {/* Wallet Card */}
        {walletInfo && (
          <div className="mx-3 mb-3">
            <div 
              className="relative overflow-hidden rounded-xl p-3.5"
              style={{ 
                backgroundColor: colors.cardBg,
                border: `1px solid ${colors.border}`
              }}
            >
              <div className="absolute top-0 right-0 w-24 h-24 bg-emerald-500/5 blur-2xl rounded-full" />
              
              <div className="relative">
                <div className="flex items-center justify-between mb-3">
                  <div className="flex items-center gap-2">
                    <div 
                      className="w-7 h-7 rounded-lg flex items-center justify-center"
                      style={{ 
                        backgroundColor: isDark ? 'rgba(16, 185, 129, 0.15)' : 'rgba(16, 185, 129, 0.1)',
                        border: isDark ? '1px solid rgba(16, 185, 129, 0.25)' : '1px solid rgba(16, 185, 129, 0.2)'
                      }}
                    >
                      <Wallet className="h-3.5 w-3.5 text-emerald-500" />
                    </div>
                    <span className="text-xs font-medium" style={{ color: colors.textSecondary }}>
                      {usingAgentWallet ? t('usermenu.agentWallet') : t('usermenu.balance')}
                    </span>
                  </div>
                  <span className="text-lg font-bold tracking-tight" style={{ color: colors.textPrimary }}>
                    {formatBalance(walletInfo.balance_usd)}
                  </span>
                </div>
                
                <div className="grid grid-cols-2 gap-3">
                  <div 
                    className="flex items-center gap-2 px-2.5 py-2 rounded-lg"
                    style={{ 
                      backgroundColor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                      border: `1px solid ${colors.border}`
                    }}
                  >
                    <TrendingUp className="h-3 w-3 text-emerald-500" />
                    <div>
                      <p className="text-[10px] uppercase tracking-wider font-medium" style={{ color: colors.textMuted }}>{t('usermenu.earned')}</p>
                      <p className="text-xs font-semibold text-emerald-500">
                        {formatBalance(
                          usingAgentWallet
                            ? (walletInfo as { total_earned_usd?: number }).total_earned_usd
                            : (walletInfo as { lifetime_earnings_usd?: number }).lifetime_earnings_usd
                        )}
                      </p>
                    </div>
                  </div>
                  <div 
                    className="flex items-center gap-2 px-2.5 py-2 rounded-lg"
                    style={{ 
                      backgroundColor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                      border: `1px solid ${colors.border}`
                    }}
                  >
                    <ArrowUpRight className="h-3 w-3 text-rose-500" />
                    <div>
                      <p className="text-[10px] uppercase tracking-wider font-medium" style={{ color: colors.textMuted }}>
                        {usingAgentWallet ? t('usermenu.spent') : t('usermenu.fees')}
                      </p>
                      <p className="text-xs font-semibold text-rose-500">
                        {formatBalance(
                          usingAgentWallet
                            ? (walletInfo as { total_spent_usd?: number }).total_spent_usd
                            : (walletInfo as { lifetime_fees_usd?: number }).lifetime_fees_usd
                        )}
                      </p>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        <DropdownMenuSeparator className="mx-3" style={{ backgroundColor: colors.border }} />

        {/* Team Switcher */}
        {teams.length > 0 && (
          <>
            <div className="px-3 py-2">
              <button
                onClick={() => setShowTeamsSubmenu(!showTeamsSubmenu)}
                className="w-full flex items-center justify-between px-3 py-2.5 rounded-xl text-sm transition-all duration-200 group"
                style={{ 
                  backgroundColor: colors.cardBg,
                  border: `1px solid ${colors.border}`
                }}
              >
                <div className="flex items-center gap-2.5">
                  <div 
                    className="w-7 h-7 rounded-lg flex items-center justify-center"
                    style={{ 
                      backgroundColor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)',
                      border: `1px solid ${colors.border}`
                    }}
                  >
                    <Building2 className="w-3.5 h-3.5" style={{ color: colors.textMuted }} />
                  </div>
                  <span className="font-medium text-xs uppercase tracking-wider" style={{ color: colors.textSecondary }}>{t('usermenu.org')}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium truncate max-w-[100px]" style={{ color: colors.textPrimary }}>
                    {currentTeam?.name || 'Personal'}
                  </span>
                  <motion.div
                    animate={{ rotate: showTeamsSubmenu ? 90 : 0 }}
                    transition={{ duration: 0.2 }}
                    className="w-5 h-5 rounded-md flex items-center justify-center"
                    style={{ backgroundColor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)' }}
                  >
                    <ChevronRight className="w-3.5 h-3.5" style={{ color: colors.textMuted }} />
                  </motion.div>
                </div>
              </button>

              <AnimatePresence>
                {showTeamsSubmenu && (
                  <motion.div
                    initial={{ height: 0, opacity: 0 }}
                    animate={{ height: 'auto', opacity: 1 }}
                    exit={{ height: 0, opacity: 0 }}
                    transition={{ duration: 0.2, ease: [0.4, 0, 0.2, 1] }}
                    className="overflow-hidden"
                  >
                    <div 
                      className="mt-2 space-y-1 pl-3 ml-3"
                      style={{ borderLeft: `2px solid ${colors.border}` }}
                    >
                      {teams.map((team) => (
                        <button
                          key={team.id}
                          onClick={() => {
                            navigate(`/teams/${team.id}`);
                            setIsOpen(false);
                          }}
                          className={cn(
                            'w-full flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-sm transition-all duration-150 border',
                            team.id === currentTeam?.id
                              ? 'border-[#FF6B35]/30'
                              : 'hover:border-transparent border border-transparent'
                          )}
                          style={team.id === currentTeam?.id ? { backgroundColor: isDark ? `${brand.primary}15` : `${brand.primary}10` } : { backgroundColor: 'transparent' }}
                        >
                          <div className={cn(
                            'w-6 h-6 rounded-md flex items-center justify-center text-xs font-bold',
                            team.id === currentTeam?.id
                              ? 'text-white shadow-lg'
                              : 'text-white/70'
                          )}
                          style={team.id === currentTeam?.id ? { backgroundColor: brand.primary, boxShadow: `0 4px 6px -1px ${brand.glow}` } : { backgroundColor: isDark ? 'rgba(255, 255, 255, 0.1)' : 'rgba(0, 0, 0, 0.1)' }}>
                            {team.name.charAt(0).toUpperCase()}
                          </div>
                          <span className={cn(
                            'truncate flex-1 text-left',
                            team.id === currentTeam?.id ? 'font-medium' : ''
                          )}
                          style={{ color: team.id === currentTeam?.id ? brand.primary : colors.textSecondary }}>
                            {team.name}
                          </span>
                          {team.id === currentTeam?.id && (
                            <Check className="w-4 h-4" style={{ color: brand.primary }} />
                          )}
                        </button>
                      ))}
                      <button
                        onClick={() => {
                          navigate(ROUTES.TEAMS);
                          setIsOpen(false);
                        }}
                        className="w-full flex items-center gap-2.5 px-2.5 py-2 rounded-lg text-sm transition-all hover:bg-opacity-5"
                        style={{ color: colors.textMuted }}
                      >
                        <div 
                          className="w-6 h-6 rounded-md flex items-center justify-center"
                          style={{ 
                            backgroundColor: isDark ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 0, 0, 0.05)',
                            border: `1px solid ${colors.border}`
                          }}
                        >
                          <Plus className="w-3.5 h-3.5" style={{ color: colors.textMuted }} />
                        </div>
                        <span>Create or join</span>
                      </button>
                    </div>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
            <DropdownMenuSeparator className="mx-3" style={{ backgroundColor: colors.border }} />
          </>
        )}

        {/* Upgrade CTA - Velocity Orange */}
        {user.plan?.toLowerCase() === 'free' && (
          <>
            <div className="mx-3 mb-3">
              <button
                onClick={() => {
                  setIsOpen(false);
                  setShowPlanModal(true);
                }}
                className="group w-full block relative overflow-hidden rounded-xl p-3.5 transition-all duration-300 text-left"
                style={{
                  background: isDark
                    ? `linear-gradient(135deg, ${brand.primary}20 0%, #FF8C4215 50%, #FF4F5E10 100%)`
                    : `linear-gradient(135deg, ${brand.primary}10 0%, #FF8C4208 50%, #FF4F5E05 100%)`,
                  border: isDark ? `1px solid ${brand.primary}40` : `1px solid ${brand.primary}30`,
                }}
              >
                <div
                  className="absolute -top-10 -right-10 w-20 h-20 blur-xl rounded-full transition-all duration-500"
                  style={{ backgroundColor: `${brand.primary}40` }}
                />
                <div
                  className="absolute -bottom-10 -left-10 w-20 h-20 blur-xl rounded-full transition-all duration-500"
                  style={{ backgroundColor: `${brand.light}30` }}
                />

                <div className="relative flex items-center gap-3">
                  <div
                    className="w-10 h-10 rounded-xl flex items-center justify-center shadow-lg transition-all duration-300 group-hover:scale-105"
                    style={{ background: brand.gradient, boxShadow: `0 10px 15px -3px ${brand.glow}` }}
                  >
                    <Crown className="w-5 h-5 text-white" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <p className="text-sm font-bold transition-colors group-hover:text-[#FF6B35]" style={{ color: colors.textPrimary }}>
                      {t('usermenu.upgradeToPro')}
                    </p>
                    <p className="text-xs mt-0.5 leading-relaxed" style={{ color: colors.textMuted }}>
                      {t('usermenu.upgradeDescription')}
                    </p>
                  </div>
                  <ChevronRight className="w-5 h-5 transition-all flex-shrink-0 group-hover:translate-x-1" style={{ color: brand.primary }} />
                </div>

                <div
                  className="absolute bottom-0 left-0 right-0 h-0.5 opacity-60"
                  style={{ background: `linear-gradient(to right, ${brand.primary}, ${brand.light}, #f59e0b)` }}
                />
              </button>
            </div>
            <DropdownMenuSeparator className="mx-3" style={{ backgroundColor: colors.border }} />
          </>
        )}

        {/* Navigation Items */}
        <div className="px-1.5 py-1 space-y-0.5">
          <DropdownMenuItem asChild>
            <Link
              to={user.username ? `/u/${user.username}` : '/profile'}
              className="flex items-center gap-3 px-3 py-2.5 text-sm rounded-lg cursor-pointer transition-all duration-150 group"
              style={{ backgroundColor: 'transparent' }}
              onClick={() => setIsOpen(false)}
            >
              <div 
                className="w-7 h-7 rounded-lg flex items-center justify-center border transition-all"
                style={{ 
                  backgroundColor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  borderColor: colors.border
                }}
              >
                <User 
                  className="w-3.5 h-3.5 transition-colors" 
                  style={{ color: colors.textMuted }}
                />
              </div>
              <span className="font-medium transition-colors" style={{ color: colors.textSecondary }}>{t('usermenu.profile')}</span>
            </Link>
          </DropdownMenuItem>

          <DropdownMenuItem asChild>
            <Link
              to={user.username ? `/u/${user.username}/settings` : ROUTES.SETTINGS}
              className="flex items-center gap-3 px-3 py-2.5 text-sm rounded-lg cursor-pointer transition-all duration-150 group"
              onClick={() => setIsOpen(false)}
            >
              <div 
                className="w-7 h-7 rounded-lg flex items-center justify-center border transition-all group-hover:border-cyan-500/40"
                style={{ 
                  backgroundColor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  borderColor: colors.border
                }}
              >
                <Settings className="w-3.5 h-3.5 group-hover:text-cyan-500 transition-colors" style={{ color: colors.textMuted }} />
              </div>
              <span className="font-medium group-hover:text-cyan-500 transition-colors" style={{ color: colors.textSecondary }}>{t('usermenu.settings')}</span>
            </Link>
          </DropdownMenuItem>

          <DropdownMenuItem asChild>
            <Link
              to={user.username ? `/u/${user.username}/settings/billing` : ROUTES.BILLING}
              className="flex items-center gap-3 px-3 py-2.5 text-sm rounded-lg cursor-pointer transition-all duration-150 group"
              onClick={() => setIsOpen(false)}
            >
              <div 
                className="w-7 h-7 rounded-lg flex items-center justify-center border transition-all group-hover:border-emerald-500/40"
                style={{ 
                  backgroundColor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                  borderColor: colors.border
                }}
              >
                <CreditCard className="w-3.5 h-3.5 group-hover:text-emerald-500 transition-colors" style={{ color: colors.textMuted }} />
              </div>
              <span className="font-medium group-hover:text-emerald-500 transition-colors" style={{ color: colors.textSecondary }}>{t('usermenu.billing')}</span>
            </Link>
          </DropdownMenuItem>

          <div className="flex items-center gap-3 px-3 py-2.5">
            <div
              className="w-7 h-7 rounded-lg flex items-center justify-center border"
              style={{
                backgroundColor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                borderColor: colors.border
              }}
            >
              <Globe className="w-3.5 h-3.5" style={{ color: colors.textMuted }} />
            </div>
            <LanguagePicker variant="ghost" showLabel={true} className="flex-1 h-8 px-0" />
          </div>
        </div>

        {isAdmin && ADMIN_DASHBOARD_URL && (
          <>
            <DropdownMenuSeparator className="mx-3" style={{ backgroundColor: colors.border }} />
            <div className="px-1.5 py-1">
              <DropdownMenuItem asChild>
                <a
                  href={ADMIN_DASHBOARD_URL}
                  className="flex items-center gap-3 px-3 py-2.5 text-sm rounded-lg cursor-pointer transition-all duration-150 group"
                  onClick={() => setIsOpen(false)}
                  rel="noopener noreferrer"
                >
                  <div 
                    className="w-7 h-7 rounded-lg flex items-center justify-center border transition-all group-hover:border-amber-500/40"
                    style={{ 
                      backgroundColor: isDark ? 'rgba(255, 255, 255, 0.03)' : 'rgba(0, 0, 0, 0.03)',
                      borderColor: colors.border
                    }}
                  >
                    <Shield className="w-3.5 h-3.5 group-hover:text-amber-500 transition-colors" style={{ color: colors.textMuted }} />
                  </div>
                  <span className="font-medium group-hover:text-amber-500 transition-colors" style={{ color: colors.textSecondary }}>{t('usermenu.adminPanel')}</span>
                  <span 
                    className="ml-auto px-1.5 py-0.5 rounded text-[10px] font-semibold"
                    style={{ 
                      backgroundColor: isDark ? 'rgba(245, 158, 11, 0.15)' : 'rgba(245, 158, 11, 0.1)',
                      color: '#f59e0b',
                      border: isDark ? '1px solid rgba(245, 158, 11, 0.25)' : '1px solid rgba(245, 158, 11, 0.2)'
                    }}
                  >PRO</span>
                </a>
              </DropdownMenuItem>
            </div>
          </>
        )}

        <DropdownMenuSeparator className="mx-3" style={{ backgroundColor: colors.border }} />

        {/* Logout */}
        <div className="p-1.5">
          <DropdownMenuItem
            onClick={handleLogout}
            className="flex items-center gap-3 px-3 py-2.5 text-sm rounded-lg cursor-pointer transition-all duration-150 group hover:bg-rose-50 dark:hover:bg-rose-500/10"
          >
            <div 
              className="w-7 h-7 rounded-lg flex items-center justify-center transition-all group-hover:bg-rose-100 dark:group-hover:bg-rose-500/20"
              style={{ 
                backgroundColor: isDark ? 'rgba(244, 63, 94, 0.1)' : 'rgba(244, 63, 94, 0.08)',
                border: isDark ? '1px solid rgba(244, 63, 94, 0.2)' : '1px solid rgba(244, 63, 94, 0.15)'
              }}
            >
              <LogOut className="w-3.5 h-3.5 text-rose-500" />
            </div>
            <span className="font-medium text-rose-500">{t('usermenu.signOut')}</span>
          </DropdownMenuItem>
        </div>
      </DropdownMenuContent>
    </DropdownMenu>

    {/* Plan Selection Modal */}
    <PlanSelectionModal
      isOpen={showPlanModal}
      onClose={() => {
        setShowPlanModal(false);
        setIsCheckoutLoading(false);
      }}
      isFree={true}
      nextTier="Starter"
      onSelectPlan={async (planId: string, priceId?: string) => {
        if (!priceId || priceId.includes('placeholder')) {
          if (planId === 'enterprise') {
            navigate('/contact');
          }
          return;
        }

        setIsCheckoutLoading(true);
        try {
          const base = window.location.origin;
          const successUrl = user?.username
            ? `${base}/u/${user.username}/settings/billing?subscription=success`
            : `${base}/settings?tab=billing&subscription=success`;
          const cancelUrl = `${base}/dashboard?subscription=cancel`;

          const { url } = await createCheckoutSession(priceId, successUrl, cancelUrl);
          window.location.href = url;
        } catch (err) {
          setIsCheckoutLoading(false);
          console.error('Checkout error:', err);
        }
      }}
      isCheckoutLoading={isCheckoutLoading}
    />
    </>
  );
}
