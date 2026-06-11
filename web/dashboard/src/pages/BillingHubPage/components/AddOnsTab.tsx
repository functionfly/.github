import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Skeleton } from '@/components/ui/skeleton';
import type { StateFabricAddOnDTO } from '@/api/billing';
import { Check, Loader2, Package, Zap } from 'lucide-react';

interface AddOnsTabProps {
  addOnCatalog: StateFabricAddOnDTO[];
  entitledAddOnIds: string[];
  onPurchase: (addonId: string) => Promise<void>;
  isLoading: boolean;
}

export function AddOnsTab({ addOnCatalog, entitledAddOnIds, onPurchase, isLoading }: AddOnsTabProps) {
  if (isLoading) {
    return (
      <div className="space-y-6">
        <Card className="ff-card-velocity">
          <CardHeader>
            <Skeleton className="h-6 w-48" />
          </CardHeader>
          <CardContent className="space-y-4">
            {[1, 2, 3, 4].map((i) => (
              <Skeleton key={i} className="h-32 w-full" />
            ))}
          </CardContent>
        </Card>
      </div>
    );
  }

  if (addOnCatalog.length === 0) {
    return (
      <div className="space-y-6">
        <Card className="ff-card-velocity">
          <CardHeader>
            <CardTitle className="font-display flex items-center gap-2">
              <Package className="h-5 w-5 text-brand-500" />
              State Fabric Add-ons
            </CardTitle>
            <CardDescription>
              Enhance your State Fabric experience with premium add-ons
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center py-8">
              <Package className="h-12 w-12 text-text-muted mx-auto mb-3" />
              <p className="text-text-muted">No add-ons available at the moment</p>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Card className="ff-card-velocity">
        <CardHeader>
          <CardTitle className="font-display flex items-center gap-2">
            <Package className="h-5 w-5 text-brand-500" />
            State Fabric Add-ons
          </CardTitle>
          <CardDescription>
            Premium add-ons to enhance your State Fabric experience
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            {addOnCatalog.map((addon) => {
              const isEntitled = entitledAddOnIds.includes(addon.id);
              return (
                <div
                  key={addon.id}
                  className={`p-4 rounded-lg border transition-colors ${
                    isEntitled
                      ? 'border-green-500/30 bg-green-500/5'
                      : 'border-border-default bg-bg-secondary'
                  }`}
                >
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <h4 className="font-semibold">{addon.name}</h4>
                      <p className="text-2xl font-bold text-brand-500 mt-1">
                        ${addon.price}
                        <span className="text-sm text-text-muted font-normal">/{addon.period}</span>
                      </p>
                    </div>
                    {isEntitled && (
                      <Badge variant="success" className="ff-badge-success">
                        <Check className="h-3 w-3 mr-1" />
                        Active
                      </Badge>
                    )}
                  </div>

                  <p className="text-sm text-text-secondary mb-4">{addon.description}</p>

                  {isEntitled ? (
                    <div className="flex items-center gap-2 p-3 rounded-lg bg-green-500/10 border border-green-500/20">
                      <Check className="h-4 w-4 text-green-400" />
                      <span className="text-sm text-green-400">
                        You have access to {addon.name}
                      </span>
                    </div>
                  ) : (
                    <Button
                      variant="outline"
                      className="w-full border-brand-500/50 text-brand-500 hover:bg-brand-500/10"
                      onClick={() => onPurchase(addon.id)}
                    >
                      <Zap className="mr-2 h-4 w-4" />
                      Subscribe
                    </Button>
                  )}
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}